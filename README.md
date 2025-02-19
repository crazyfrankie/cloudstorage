# ZCloud
本项目是一个使用 Go 语言编写的简易云盘系统。
***
技术栈: grpc + gin + gorm + MySQL + Redis + Minio + Kafka + Prometheus + etcd
***
# 项目结构
1. 项目概览
``` 
. cloudstorage 
├── app               // 各个微服务
│   ├── file          // 文件模块
│   ├── gateway       // gin 网关
│   ├── sm            // 短信模块
│   └── user          // 用户模块
├── idl               // proto 文件
└── rpc_gen           // grpc 接口代码
```
2. file 模块
```
.file
├── biz
│   ├── repository         // 数据存储层
│   │    ├── cache         // 缓存层
│   │    └── dao           // 数据访问层
│   └── service            // 业务逻辑
├── config 
│   ├── config.go          // 配置读取
│   └── test               // test 环境
│        └── config.yaml   // 配置文件
├── ioc
│   ├── wire.go           
│   └── wire_gen.go        // 依赖注入
├── main.go                // 入口文件
├── mws 
│   └── minio.go           // minio 中间件
└── rpc
    └── server.go          // grpc server 抽象 
```
3. user 模块
```
.user
├── biz
│   ├── repository         // 数据存储层
│   │    └── dao           // 数据访问层
│   └── service            // 业务逻辑
├── config 
│   ├── config.go          // 配置读取
│   └── test               // test 环境
│        └── config.yaml   // 配置文件
├── ioc
│   ├── wire.go           
│   └── wire_gen.go        // 依赖注入
├── main.go                // 入口文件
├── mws 
│   └── jwt.go             // jwt 中间件
└── rpc
    └── server.go          // grpc server 抽象 
```
4. sm 模块
```
.sm
├── biz
│   ├── repository         // 数据存储层
│   │    ├── cache         // 缓存层
│   └── service            // 业务逻辑
├── config 
│   ├── config.go          // 配置读取
│   └── test               // test 环境
│        └── config.yaml   // 配置文件
├── ioc
│   ├── wire.go           
│   └── wire_gen.go        // 依赖注入
├── main.go                // 入口文件
└── rpc
    └── server.go          // grpc server 抽象 
```
5. gateway 网关模块
```
. gateway
├── api                    // http 处理层
│   ├── file.go            // file 模块
│   └── user.go            // user 模块
├── common
│   ├── consts             // 常量定义
│   ├── response           // http 统一出口
│   └── util               // 工具包
├── ioc
│   ├── client.go          // grpc client 初始化
│   ├── wire.go            
│   └── wire_gen.go        // 依赖注入
├── main.go
└── mws
    └── auth.go            // 身份校验中间件
```
# 项目配置文件
`.env` : 采用以下格式，且必须放置在 file & user 包下
```
MYSQL_USER=your_user
MYSQL_PASSWORD=your_password
MYSQL_HOST=your_host
MYSQL_PORT=your_port
MYSQL_DB=your_db
```
`config.yaml` : 大致采用以下格式，根据不同模块 `config.go` 的需求进行更改
```
server:
  addr: "your_addr"

mysql:
  dsn: "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"

etcd:
  addr: "your_addr"
```