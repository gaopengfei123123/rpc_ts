package services
import (
	"github.com/astaxie/beego/logs"
	"fmt"
)

// 进行 log 日志生成方案的配置
func init(){
	logs.Async(1e3)
	config := fmt.Sprintf(`{"filename":"%s"}`, LogPath )
	logs.SetLogger(logs.AdapterFile, config)
}
// redis 的一些配置
const (
	RedisHost = "localhost:6379"
	RedisPassword = ""
	RredisDB = 0
	RedisCacheTTL = 7200
)


// kafka 的一些配置
const (
	// KafkaTopic 指定的主题名
	KafkaTopic = "my-topic"
	// KafkaPatition 指定分区
	KafkaPatition = 0
	// KafkaHost kafka 链接地址
	KafkaHost = "localhost:9092"
)

// 消费者服务的一些配置
const (
	// MaxExecNum 最大执行次数
	MaxExecNum = 3
	// MaxExecTime 最大执行时间(单位 秒)
	MaxExecTime = 20
)


// log 的一些设置
const (
	// LogPath  日志地址
	LogPath = "./logs/rpc_ts.log"
)