package services

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