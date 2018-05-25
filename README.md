用于微服务的事务消息处理的中间件,暂时使用 http 通信,后期集成 grpc


### 环境
数据库配置:
```sql
database:   go
table:      rpc_ts
表结构:
CREATE TABLE `rpc_ts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `payload` text NOT NULL,
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '0=>未执行, 1=>执行中,2=>事务执行完毕,20=>事务正常取消,21=>存在非正常取消(cancel 失败)',
  `exec_num` int(11) NOT NULL DEFAULT '0' COMMENT '执行册数',
  `create_at` int(11) DEFAULT '0' COMMENT '创建时间',
  `update_at` int(11) DEFAULT '0' COMMENT '更新时间',
  `error_info` varchar(2000) NOT NULL DEFAULT '' COMMENT '错误原因',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=38 DEFAULT CHARSET=utf8;

调整数据库链接信息 file path: @rpc_ts/services/tools.go:27 GetDb() 部分调整
```

rabbitmq 配置:
```go
// file path: @rpc_ts/tools/mqhandler/rabbitmq/rabbitmq_server.go
const(
	RabbitmqHost = "amqp://guest:guest@localhost:5672/"
	QueueName = "hello"
	ConsumerName = ""
	Exchange = "rpc_transaction"
	Durable = false
	DeleteWhenUnused = false
	Exclusive = false
	NoWait = false
	AutoAck = true
	NoLocal = false
	Mandatory = false
	Immediate = false
	DelayExpiration = "5000" // 设置5秒的队列过期时间, 这里仅仅用在延时队列设置当中
)
```

http server 配置:
host: http://127.0.0.1:8899
file path: @rpc_ts/main.go

LOG 配置
```go
// file path: @rpc_ts/tools/loghandler/handler.go
const (
	// LogPath  日志文件地址
	LogPath = "./logs/rpc_ts.log"
	// log 类型, 支持 console, file
	ConsoleType = "console"
)
```


### 启动
依赖:
    - mysql
    - rabbitmq
```bash
git clone xxx@xxx.git  rpc_ts
cd rpc_ts
govendor sync
go run main.go   #或编译一下 go build main.go && ./main
```

### 测试
在 **postman** 中导入 `@rpc_ts/rpc_ts.postman_collection.json`,进行测试