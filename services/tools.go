// Package services 这里放置一一些公用的组件
package services

import(
	// "context"
	// "time"
	"github.com/astaxie/beego/logs"
	"database/sql"
	_ "github.com/GO-SQL-Driver/MySQL"	// 引入 mysql 驱动
	"encoding/json"
	mq "rpc_ts/tools/mqhandler"
)

// MQTemplate 存储的消息模型
type MQTemplate struct{
	ID int `json:"ID"`				// 消息在数据库上的 id
	ExecTime int `json:"exec_time"`	//上次执行时间
}
// GetMQServer 获取 mq 服务实体(kafka 和 rabbitmq 可选)
func GetMQServer() mq.MQ{
	var mqHandler mq.MQ
	mqHandler = new(mq.MQService)
	return mqHandler
}

// GetDb 获取 mysql 链接信息
func GetDb() *sql.DB{
	db, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:33060)/go?charset=utf8")

	return db
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

	logs.Info("打印队列:", JSONToStr(tpl))
	var server ServerForm
	if err != nil {
		return server, err
	}

	json.Unmarshal([]byte(rpcTs.Payload), &server)
	server.ID = tpl.ID
	return server, nil
}