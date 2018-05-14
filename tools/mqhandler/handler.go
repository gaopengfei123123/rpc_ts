package mqhandler

import(
	// "fmt"
	// "rpc_ts/tools/mqhandler/kafka"
	"rpc_ts/tools/mqhandler/rabbitmq"
)


// MQService mq 服务,目前暂定 kafka, 包含 send 和 read 两个方法
type MQService struct{
	// kafka.Kafka
	rabbitmq.Rabbitmq
}
// MQ 队列应该实现的方法
type MQ interface{
	Read(func(jsonStr []byte))
	Send(key string,value string)
	Delay(key string,value string, expire string)
}