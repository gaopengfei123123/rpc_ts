package router
import (
	"github.com/gin-gonic/gin"
	"rpc_ts/controllers"
	"net/http"
	logs "rpc_ts/tools/loghandler"
	"fmt"
	"encoding/json"
	"strconv"
)


// RegistRouter 注册路由
func RegistRouter(r *gin.Engine) *gin.Engine {
	// 指定访问的静态文件
	r.StaticFile("/", "./view/index.html")
	// 指定访问的目录
	r.StaticFS("/static", http.Dir("./view/static"))

	r.GET("/hello", controllers.HelloPage)

	// 接收执行事务的接口  Serial(串行)  Parallel(并行)
	r.POST("/client", controllers.Client)

	// 完全成功的接口
	r.POST("/api/test/try", trySuccess)
	r.POST("/api/test/confirm", confirmSuccess)
	r.POST("/api/test/cancel", cancelSuccess)

	// try 中存在失败的接口
	r.POST("/api/test_try_false/try", tryFault)
	

	// commit 中存在失败但可以回滚成功的接口
	r.POST("/api/test_confirm_fault/try", trySuccess)
	r.POST("/api/test_confirm_fault/confirm", confirmFault)
	r.POST("/api/test_confirm_fault/cancel", cancelSuccess)


	// 执行 cancel 回滚不成功的接口
	r.POST("/api/test_cancel_fault/try", trySuccess)
	r.POST("/api/test_cancel_fault/confirm", confirmSuccess)
	r.POST("/api/test_cancel_fault/cancel", cancelFault)


	// 执行 cancel 回滚超时的接口
	r.POST("/api/test_cancel_timeout/try", trySuccess)
	r.POST("/api/test_cancel_timeout/confirm", confirmSuccess)
	r.POST("/api/test_cancel_timeout/cancel", cancelTimeout)

	// 串行执行 step_1 完全成功的例子
	r.POST("/api/step/try", step1Try)
	r.POST("/api/step/confirm", step1Confirm)
	r.POST("/api/step/cancel", step1Cancel)

	// // 串行执行 step_2 完全成功的例子
	// r.POST("/api/step_2/try", step2Try)
	// r.POST("/api/step_2/confirm", step2Confirm)
	// r.POST("/api/step_2/cancel", step2Cancel)

	// // 串行执行 step_3 完全成功的例子
	// r.POST("/api/step_3/try", step3Try)
	// r.POST("/api/step_3/confirm", step3Confirm)
	// r.POST("/api/step_3/cancel", step3Cancel)





	return r
}

// try成功
func trySuccess(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	statusCode = 200

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
		Params string `json:"params"`
		ExParams string `json:"ex_params"`
		
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {

		logs.Error(testForm, "收到的数据_client")
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}

// try 失败
func tryFault(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	errorCode = 401
	errorMsg = "something was wrong"

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}

// confirm 成功
func confirmSuccess(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	statusCode = 200

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}

// confirm 失败
func confirmFault(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	errorCode = 401
	errorMsg = "something was wrong"

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}


func cancelSuccess(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	statusCode = 200

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "cancel api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}


// cancel 失败
func cancelFault(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	errorCode = 401
	errorMsg = "something was wrong"

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(200, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}

// cancel 超时
func cancelTimeout(c *gin.Context){
	var statusCode int
	var errorCode int
	var errorMsg string
	errorCode = 408
	errorMsg = "something was wrong"

	var testForm struct{
		Message string `json:"message"`
		Type	string	`json:"type"`
		Code	int		`json:"code"`
		ErrorCode int	`json:"error_code"`
		ErrorMessage string	`json:"error_message"`
		Data string `json:"data"`
	}
	// 验证是否成功绑定
	if c.BindJSON(&testForm) == nil {
		testForm.Type = "try api"
		testForm.Code = statusCode
		testForm.ErrorCode = errorCode
		testForm.ErrorMessage = errorMsg
		testForm.Data = "{\"message\": \"try\"}"
		c.JSON(408, testForm)
	} else {
		// 处理失败时的返回
		c.JSON(408, gin.H{
			"error_code": 408,
			"error_message": "time out",
		})
	}
}


// 接收请求的 json 格式
type testForm struct{
	Params		int 	`json:"params"`		// 获取原始参数
	ExParams  	int		`json:"ex_params"`	// 获取额外参数
}

// Forms 接收请求的 json 格式
type Forms struct{
	Params		string 	`form:"params" json:"params"`		// 获取原始参数
	ExParams  	string	`form:"ex_params" json:"ex_params"`	// 获取额外参数
}

func decodeParams(c *gin.Context) (Forms, error){
	var form Forms
	buf := make([]byte, 1024)  
	n, _ := c.Request.Body.Read(buf)  
	err := json.Unmarshal(buf[0:n], &form)
	fmt.Printf("当前解析参数: %s \n",string(buf[0:n]))
	return form, err
}
// 串行 step_1 的 try 成功
func step1Try(c *gin.Context){
	logs.Debug("step Try 开始执行")

	form, err := decodeParams(c)

	if err == nil {
		logs.Error("接收参数 prams: %s, ex_params: %s", form.Params, form.ExParams)

		var res int
		a , _ := strconv.Atoi(form.Params)
		b , _ := strconv.Atoi(form.ExParams)

		res = a + b
		// 执行失败条件
		if res > 10 {
			c.JSON(200, gin.H{
				"code": 200,
				"error_code": 401,
				"error_message": "result must <= 10",
			})
		// 制造一种超时条件
		} else if res == 10 {
			c.JSON(408, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		// 执行成功
		} else {
			c.JSON(200, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		}
	} else {
		logs.Error("解析参数失败")
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing 233" + err.Error(),
		})
	}
}
// 串行 step_1 的 confirm 成功
func step1Confirm(c *gin.Context){
	logs.Debug("step Confirm 开始执行")

	form, err := decodeParams(c)

	if err == nil {
		logs.Error("接收参数 prams: %s, ex_params: %s", form.Params, form.ExParams)

		var res int
		a , _ := strconv.Atoi(form.Params)
		b , _ := strconv.Atoi(form.ExParams)

		res = a  + b
		// 执行失败条件
		if res > 10 {
			c.JSON(200, gin.H{
				"code": 200,
				"error_code": 401,
				"error_message": "result must <= 10",
			})
		// 制造一种超时条件
		} else if res == 10 {
			c.JSON(408, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		// 执行成功
		} else {
			c.JSON(200, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		}
	} else {
		logs.Error("解析参数失败")
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing 233" + err.Error(),
		})
	}
}
// 串行 step_1 的 cancel 
func step1Cancel(c *gin.Context){
	logs.Debug("step Cancel 开始执行")

	form, err := decodeParams(c)
	
	if err == nil{
		logs.Debug("接收参数 prams: %d, ex_params: %d", form.Params, form.ExParams)

		var res int
		a , _ := strconv.Atoi(form.Params)
		b , _ := strconv.Atoi(form.ExParams)
		res = b - a

		// 执行失败条件
		if res > 5 {
			c.JSON(200, gin.H{
				"code": 200,
				"error_code": 401,
				"error_message": "result must <= 5",
			})
		// 制造一种超时条件
		} else if res == 5 {
			c.JSON(408, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		// 执行成功
		} else {
			c.JSON(200, gin.H{
				"data": strconv.Itoa(res),
				"code": 200,
			})
		}
	} else {
		// 处理失败时的返回
		c.JSON(400, gin.H{
			"error_code": 401,
			"error_message": "get nothing",
		})
	}
}