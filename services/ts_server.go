//Package services 这是处理事务任务的主体,通过监听队列来获取事务处理任务,目前支持同步执行和串行执行
// 同步执行采用的并发处理机制,每个都是独立的
package services

import(
	"fmt"
	"encoding/json"
	// "sync"
	"context"
	"time"
	"net/http"
	"io/ioutil"
	"bytes"
	logs "rpc_ts/tools/loghandler"
	// 引入 mysql 驱动
	"database/sql"
	_ "github.com/GO-SQL-Driver/MySQL" // 引入 mysql 驱动
	"rpc_ts/tools/uniqueid"
	"strconv"
)

const(	
	// syncMode 同步操作模式
	syncMode = "sync"		
	// asyncMode 异步操作模式
	asyncMode = "async"		
	// queueExpire 延时队列的延时执行时间,单位毫秒
	queueExpire = "10000"
)
// 用于生成全局唯一 id 的结构体
var uniqueWorker *uniqueid.Worker

func init(){
	// 初始化的时候加载计数实例
	uniqueWorker, _ = uniqueid.NewWorker(1)
}



// ServerForm 接收参数时的 json 格式
type ServerForm struct{
	Type string `json:"type" binding:"required"`
	Task []ServerItem
	ExecNum int `json:"exec_num"`		//执行次数
	ExecTime int `json:"exec_time"`		//执行时间
	ID	int 	`json:"ID"`				//数据库主键
	Step	int `json:"step"`			//执行的步骤 并行(async)时,0=>try;1=>commit;2=>cancel;   串行(sync)时,代表执行到第几条命令
	ErrorMsg string `json:"error_msg"`  //执行出错的原因
}

// ServerItem 单个任务需要的结构
type ServerItem struct{
	API string `json:"api" binding:"required"`				// 需要执行的 api
	Try string `json:"try" binding:"required"`				// try 请求时的参数
	Confirm string `json:"confirm" binding:"required"`		// confirm 请求的参数
	Cancel string `json:"cancel" binding:"required"`		// cancel 请求时的参数
	Step int `json:"step"`									// 单位操作执行阶段 0 为 try 阶段,1为 commit 阶段,2为 cancel 阶段
	TryStatus string	`json:"try_status"`					// 各阶段的执行任务 有 wait,done,false 三种情况
	CommitStatus string `json:"commit_status"`
	CancelStatus string `json:"cancel_status"`
	ExParams string `json:"ex_params"`						// 来自上次请求时的参数, 仅在串行执行的时候使用
	Response string `json:"response"`						// 执行 confirm 接口时返回的参数 用于串行时的
}

// ServerService 用于处理队列任务的模块
func ServerService(jsonStr []byte){

	// 每次最开始接受处理的时候设置一个唯一ID (目前使用的是 snowflake 机制)
	uid := uniqueWorker.GetID()
	logs.SetUniqueID(uid)
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
		}
	}()

	// 将消息转换成结构
	mqTpl := MQTemplate{}
	json.Unmarshal(jsonStr, &mqTpl)

	
	requestForm, err := mqTpl.searchInDB()
	checkErr(err)

	switch requestForm.Type {
	case syncMode:
		// 顺序执行
		syncHandler(requestForm)
	case asyncMode:
		// 并发执行
		asyncHandler(requestForm)
	default:
		logs.Error(requestForm, "操作异常")
	}
}



// 顺序执行
func syncHandler(req ServerForm){
	logs.Info("开始执行串行操作")

	// 执行次数加1
	req.ExecNum ++
	// 检测执行次数
	if req.ExecNum > MaxExecNum {
		logs.Error("已超过最大执行次数,", req.toString())
		return 
	}


	// 用于承接每次请求调用回的参数
	var exParams string
	var status int		// 1 代表执行成功, 2代表超时,需要延时执行, 3 代表需要执行 cancel流程
	exParams = "0"
	for startIndex := req.Step; startIndex < len(req.Task); startIndex++ {
		req.Task[startIndex].ExParams = exParams
		exParams, status  = req.execSingleTask(startIndex)

		if status == 1 {
			// 将返回信息存入当前任务的 Response, 并传递给下一个任务的 exParams 中
			req.Task[startIndex].Response = exParams
			req.Step++
			req.updateStatus()
			logs.Error("index: %d task 返回内容为: %s", startIndex, exParams)
		} else {
			logs.Error("index: %d 执行失败了 ,状态码: %d", startIndex, status)
		}
	} 

	logs.Debug("这里是同步操作")
}


// 自行单个任务,返回响应结果
func (req *ServerForm) execSingleTask(index int) (response string, status int){
	var task ServerItem
	task = req.Task[index]

	var res respBody

	resChan := make(chan respBody, 1)
	defer close(resChan)

	if (task.TryStatus == "") {
		execItem("try", index, task, resChan)
		res = <- resChan
		logs.Error("收到 try 消息_server")
		logs.Debug(res)

		if ( res.Status == 200 ) {
			execItem("confirm", index, task, resChan)
			res = <- resChan
			logs.Error("收到 confirm 消息_server")
			logs.Debug(res)

			if res.Status == 200 {
				logs.Debug("task %d 执行完毕,返回内容 $s", index, res.Body)
				return res.Body, 1
			}
		}

		return res.Error, 0
	}




	return "execSingleTask error", 1
}

// 并发执行
func asyncHandler(req ServerForm){
	logs.Info("开始执行并发操作")
	// 执行次数加1
	req.ExecNum ++
	// 检测执行次数
	if req.ExecNum > MaxExecNum {
		logs.Error("已超过最大执行次数,", req.toString())
		return 
	}
	
	// 根据不同的执行步骤进行操作
	switch req.Step {
	case 0:
		req.combineTry()
	case 1:
		req.combineCommit()
	case 2:
		req.combineBreak()
	}
}



// 同步执行 try 步骤
func (req *ServerForm) combineTry(){
	// 创建对应数量的通信通道接收消息
	resChan := make(chan respBody, len(req.Task))
	defer close(resChan)

	logs.Info("执行 try 操作")

	// 遍历任务列表, 并发的进行 try 请求
	for index, value := range req.Task{
		go execItem("try", index, value, resChan)
	}

	// 判断整体的 try 是否通过
	isPass := true
	// 将异步的数据导出
	var result []respBody 
	for i:= 0; i < len(req.Task); i++ {
		res := <-resChan
		
		if res.Status != 200 {
			isPass = false
		}

		result = append(result, res)
	}
	
	// 如果不通过的话
	if !isPass {
		logs.Info("try 操作返回错误,执行 cancel 阶段")
		req.cancel(result)
	} else {
		req.Step ++
		logs.Debug("try 操作成功, 持久化后进入 commit 阶段")
		// 更新数据库信息
		req.updateStatus()
		// 执行并发提交
		req.combineCommit()
	}
}

// 执行单个请求操作, 可根据 taskType 来区分请求的包体
func execItem(taskType string, index int,task ServerItem,  resultChan chan<- respBody){
	// 类似 try catch 的一个防报错机制
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
		}
	}();
	logs.Info("请求开始, API: %s, type: %s", task.API, taskType)

	// 创建一个阻塞通道,用于执行请求任务,也是为了计算超时时间
	dst := make(chan respBody)
	defer close(dst)

	// 这里是进行 post 请求的主体
	go func(task ServerItem, taskType string) {
		start := time.Now()
		var resp respBody
		var url string
		switch taskType {
		case "try":
			url = fmt.Sprintf("%s/try", task.API)
			resp = postClien(url, task.Try, task.ExParams)
		case "confirm":
			url = fmt.Sprintf("%s/confirm", task.API)
			resp = postClien(url, task.Confirm, task.ExParams)
		case "cancel":
			url = fmt.Sprintf("%s/cancel", task.API)
			// 当执行取消操作的时候将会把 confirm 操作的返回值当做额外参数放进去
			resp = postClien(url, task.Cancel, task.Response)
		}
		// 对返回内容打上 api 信息
		resp.API = url
		
		ttl := time.Since(start)

		resp.ExecTime =  strconv.FormatFloat(ttl.Seconds() * 1000, 'f', 3, 64) + " ms"
		logs.Info(resp, "请求完毕")

		// 将请求结果导出
		dst <- resp

	}(task, taskType)

	// 这里就是声明一个倒计时
	ctx, cancel := context.WithTimeout(context.Background(),MaxExecTime * time.Second)
	defer cancel()

	// 监听超时时间
	LOOP:
	for {
		select {
		case resp := <-dst:
			// 当请求正常时直接怼到结果信道中
			resp.Index = index
			resultChan <- resp
			break LOOP
		case <-ctx.Done():
			// 当请求超时时,需要生成 log 并返回错误信息
			errResp := respBody{
				API: task.API,
				Status: 400,
				Body: "",
				Error: "exec timeout",
				Index: index,
				ErrorCode: 408,
			}
			resultChan <- errResp

			logs.Warn(errResp)
			break LOOP
		}
	}

}
// 同步执行 commit 步骤
// 思路:
// 1. 首先过滤出来未执行和执行失败的任务下标,准备进行处理
// 2. 进行批处理执行
// 3. 根据执行请求的返回来进行操作: 1>如果成功那么做标记后返回内容,2> 如果失败: 一如果出现执行超时情况则标记失败准备下次请求,如果是接口返回无法提交或500则整体执行 cancel 操作
func (req *ServerForm) combineCommit(){
	// 首先筛选出来未执行的任务的下标, 因为执行到这一步的时候已经是 try 操作完毕了
	// 因此这里主要看的是 ServerForm 的Step 和 Item 当中的 commit 的执行状态
	logs.Debug("开始执行 commit 部分")

	resChan := make(chan respBody, len(req.Task))
	defer close(resChan)

	// 装载合法数据的索引, 标明哪几条任务再执行,然后也是接收这几条任务的 channel,以免发生漏消息或者阻塞
	var indexFilter []int
	for index, value := range req.Task {
		if value.CommitStatus != "done" {
			indexFilter = append(indexFilter, index)
			// 单个执行提交操作
			go execItem("confirm", index, value, resChan)
		}
	}

	// 判断整体的 confirm 是否通过(如果部分未通过则会重回队列)
	isPass := true
	// 判断是否产生需要终止整个事务的事件
	isBreak := false
	// 将异步的数据导出
	var result []respBody 

	for i:= 0; i < len(indexFilter); i++ {
		res := <-resChan
		if res.Status != 200 {
			// 只要请求不成功就不能通过,但是还需要判断是否终止操作
			isPass = false
			// 执行超时的则重回队列
			if res.ErrorCode != 408 {
				// 出现非超时错误时将回滚整个任务列表
				isBreak = true
			}
			req.Task[res.Index].CommitStatus = "false"
		} else {
			req.Task[res.Index].CommitStatus = "done"
		}
		result = append(result, res)
	}

	// confirm 执行出现失败, 回滚已执行的任务
	if isBreak {
		logs.Info("confirm 执行出现失败, 回滚已执行的任务")
		req.Step++
		req.updateStatus()
		req.combineBreak()
		return
	}

	if !isPass {
		// 执行重回队列操作
		logs.Debug("执行重回队列的操作 commit step")
		req.insertMQ()
		return
	}

	// 完成动作
	logs.Info("任务执行完毕, 准备持久化数据")
	req.success()
}

// 执行单个 commit 操作
func execCommit(task ServerItem, resChan chan<- respBody, ){
	// 类似 try catch 的一个放报错机制
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
		}
	}()

	// 创建一个阻塞通道,用于执行请求任务,也是为了计算超时时间
	dst := make(chan respBody)
	defer close(dst)
}


func (req *ServerForm) toString() string{
	jsonByte, _ := json.Marshal(req)
	return string(jsonByte)
}


// 更新当前任务的mysql状态
func (req *ServerForm) updateStatus(){
	logs.Info("持久化状态,ID: %d", req.ID)
	logs.Debug(req)
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")
	sql := "UPDATE rpc_ts SET payload=?,status=1,exec_num=?,update_at=?,error_info=? WHERE id=?"
	stmt, err := db.Prepare(sql)
	checkErr(err)

	_, err = stmt.Exec(req.toString(), req.ExecNum, time.Now().Unix(), req.ErrorMsg, req.ID)
	checkErr(err)
	logs.Info(" 持久化成功,ID: %d", req.ID)
}

// 并行操作 try 不通过直接取消事务,
func (req *ServerForm) cancel(errMsg []respBody){
	logs.Error("准备开始取消")
	errStr := JSONToStr(errMsg)
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")
	sql := "UPDATE rpc_ts SET payload=?, status=20, exec_num=?, update_at=?, error_info=? WHERE id=?"
	stmt, err := db.Prepare(sql)
	checkErr(err)
	_, err = stmt.Exec(req.toString(), req.ExecNum, time.Now().Unix(), errStr, req.ID)
	checkErr(err)
	logs.Info("任务已取消, ID: %d: ",req.ID)
	logs.Error("此处应该向某处发送 [事务取消] 通知")
}

// 存在非正常取消操作
func (req *ServerForm) crash(errMsg []respBody){
	errStr := JSONToStr(errMsg)
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")
	sql := "UPDATE rpc_ts SET payload=?, status=21, exec_num=?, update_at=?, error_info=? WHERE id=?"
	stmt, err := db.Prepare(sql)
	checkErr(err)
	_, err = stmt.Exec(req.toString(), req.ExecNum, time.Now().Unix(), errStr, req.ID)
	checkErr(err)
	logs.Error("此处应该向某处发送 [事务异常] 通知")
}

// 事务执行完成动作
func (req *ServerForm) success(){
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")
	sql := "UPDATE rpc_ts SET payload=?, status=2, exec_num=?, update_at=? WHERE id=?"
	stmt, err := db.Prepare(sql)
	checkErr(err)
	_, err = stmt.Exec(req.toString(), req.ExecNum, time.Now().Unix(), req.ID)
	checkErr(err)
	logs.Info("已向数据库标记执行成功")
}

// 插入MQ
func (req *ServerForm) insertMQ() {
	if (req.ExecNum >= MaxExecNum){
		logs.Error("超过最大执行次数 3 次")

		var respbd []respBody
		respbd = append(respbd, respBody{Error: "out of max exec times",ErrorCode: 409,})
		req.cancel(respbd)
		return
	}

	jsonStr := MQTemplate{
		ID: req.ID,
		ExecTime: int(time.Now().Unix()),
	}

	insertKey := fmt.Sprintf("ts_queue_%d_%d", req.ID, req.ExecNum)
	logs.Info("插入队列, 主键ID: %d", req.ID)
	// 向消息队列中发送消息
	mq := GetMQServer()
	mq.Delay(insertKey,JSONToStr(jsonStr), queueExpire)
}

// 执行中断事务的操作
func (req *ServerForm) combineBreak(){
	logs.Info("进入 cancel 阶段")

	resChan := make(chan respBody, len(req.Task))
	defer close(resChan)

	// 装载合法数据的索引, 标明哪几条任务再执行,然后也是接收这几条任务的 channel,以免发生漏消息或者阻塞
	var indexFilter []int
	for index, value := range req.Task {
		if value.CommitStatus == "done" {
			indexFilter = append(indexFilter, index)
			// 单个执行取消操作
			go execItem("cancel", index, value, resChan)
		}
	}

	// 将异步的数据导出
	var result []respBody 
	// 判断整体的 cancel 是否通过(如果部分未通过则会重回队列)
	isPass := true
	// 判断是否产生需要终止整个事务的事件
	isBreak := false
	for i:= 0; i < len(indexFilter); i++ {
		res := <-resChan
		if res.Status != 200 {
			// 当执行cancel 都出现错误的时候说明回滚不成功,
			isPass = false
			// 执行超时的则重回队列
			if res.ErrorCode != 408 {
				isBreak = true
			}
			req.Task[res.Index].CancelStatus = "false"
		} else {
			req.Task[res.Index].CancelStatus = "done"
		}
		result = append(result,res)
	}

	//更新一下状态
	req.updateStatus()

	if isBreak {
		// 当存在事务无法回滚的情况
		req.crash(result)
		return
	}

	if !isPass {
		// 执行重回队列操作
		logs.Info("执行重回队列的操作 cancel step")
		req.insertMQ()
		return
	}

	// 全部正常回滚则关闭事务
	req.cancel(result)
	logs.Info("终止操作完毕!ID:", req.ID)
}


// ==================================TOOLS===================================

// JSONToStr 转字符串的统一方法
func JSONToStr(req interface{}) string{
	strByte, _ := json.Marshal(req)
	return string(strByte)
}


// 统一的返回格式内容
// ```json
// {
// 	"code": 200,	
// 	"error_code": 400,
// 	"error_message": "错误信息",
// 	"data": "返回内容的 json 字符串"
// }

// ```

// 进行 post 请求时的返回的结构体
type respBody struct{
	Status int		`json:"code"`			// 请求任务后返回的状态码
	Body string		`json:"data"`			// 返回的信息主体
	Error string	`json:"error_message"`	// 返回的错误信息
	API string								// 请求的 api
	Index int								// 任务所属的下标
	ErrorCode int 	`json:"error_code"`		// 错误编码
	ExecTime  string 						// 请求执行时间
}


// 向远程请求时发送的格式
type postBody struct{
	Params string 	`json:"params"`		// 请求时的参数
	ExParams string `json:"ex_params"` 	// 请求时的额外参数, 当 type=sync 的时候会把上个任务的返回值传到这里
}
// post 请求工具
func postClien(url string, jsonStr string, exParams string) respBody {
	// 类似 try catch 的一个放报错机制
	defer func() {
		if err := recover(); err != nil {
			logs.Error(err)
		}
	}()

	postBody := postBody{
		Params: jsonStr,
		ExParams: exParams,
	}

	jsonByte, err := json.Marshal(postBody)
	checkErr(err)

	logs.Error(string(jsonByte), "即将发送的数据")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()


	body, _ := ioutil.ReadAll(resp.Body)

	// 如果是正常返回则解析 json 数据
	if resp.StatusCode == 200 {
		respForm := respBody{}
		json.Unmarshal(body, &respForm)

		if (respForm.ErrorCode != 0) {
			respForm.Status = 401
		}
		
		return respForm
	} 
	// 非200的请求统统属于
	// 返回一个指定的响应结构
	return respBody{
		API: url,
		Status: resp.StatusCode,
		Body: string(body),
		Error: "request error",
		ErrorCode: resp.StatusCode,
	}
	
}