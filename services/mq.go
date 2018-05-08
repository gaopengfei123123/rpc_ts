package services

import(
	"github.com/segmentio/kafka-go"
	"fmt"
	"context"
	// "time"
	"github.com/astaxie/beego/logs"
	"database/sql"
	_ "github.com/GO-SQL-Driver/MySQL"	// 引入 mysql 驱动
	"encoding/json"
)

func init(){
	fmt.Println("初始化 log 配置")
	// log 开异步
	logs.Async(1e3)
	config := fmt.Sprintf(`{"filename":"%s","separate":["error", "warning", "notice", "info", "debug"]}`, LOG_PATH )
	logs.SetLogger(logs.AdapterMultiFile, config)
}

// MQService mq 服务,目前暂定 kafka, 包含 send 和 read 两个方法
type MQService struct{}


// MQTemplate 存储的消息模型
type MQTemplate struct{
	ID int `json:"ID"`				// 消息在数据库上的 id
	ExecTime int `json:"exec_time"`	//上次执行时间
}

//从数据库中搜索数据
func (tpl *MQTemplate) searchInDB() (ServerForm, error){

	var rpcTs struct{
		ID int
		Payload string
		ExecNum int
		CreateAt int
		UpdateAt int
	}
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")

	// 只查询可运行的任务状态
	err := db.QueryRow(`SELECT id,payload,exec_num,create_at,update_at FROM rpc_ts WHERE id = ? AND status<2 AND update_at<=?`, tpl.ID, tpl.ExecTime).Scan(&rpcTs.ID,&rpcTs.Payload,&rpcTs.ExecNum,&rpcTs.CreateAt,&rpcTs.UpdateAt)

	logs.Error("打印队列:", JSONToStr(tpl))
	var server ServerForm
	if err != nil {
		return server, err
	}

	json.Unmarshal([]byte(rpcTs.Payload), &server)
	server.ID = tpl.ID
	return server, nil
}

// Send 向队列插入数据 (目前是使用 kafka)
func (mq *MQService) Send(key string,value string){
	conn, err := kafka.DialLeader(context.Background(), "tcp", KAFKA_HOST, KAFKA_TOPIC, KAFKA_PARTITION)
	checkErr(err)
	defer conn.Close()

	conn.WriteMessages(
		kafka.Message{
			Key: []byte(key),
			Value: []byte(value),
		},
	)


	logs.Debug("already send msg, key:", key)
}

func (mq *MQService) Read(){
	logs.Debug("mq reader is starting")
	startReading()
}

// 开始监听消息队列
func startReading(){
	// make a new reader that consumes from topic-A, partition 0, at offset 42
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{KAFKA_HOST},
		Topic:     KAFKA_TOPIC,
		Partition: KAFKA_PARTITION,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})


	ctx := context.Background()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			break
		}
		logs.Info("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
		go ServerService(m.Value)

		r.CommitMessages(ctx, m)
	}
	defer r.Close()
}