## 代码结构

``` go
├── build-ci.sh      // 编译及unittest执行
├── common           // 公用的方法包，同时包含const/global 全局的变量
├── conf             // 配置包
├── env              // 全局环境包，包含conf/logger/mysqlDB/Mailer 初始化的对象
├── html             
├── libs             // 公共类库，包含logger，mail类
│   ├── logger
│   └── mail
├── main.go
├── runtime          // 测试的配置，线上配置放到了deploy/conf中
├── schema           // 数据表schema 
├── server           // 服务业务层 
│   ├── dao          // 数据访问层 
│   ├── http_server  // http服务层
│   │   ├── api      // 所有api的逻辑放到了api包
│   │   └── gin.go   // gin启动及路由配置
│   ├── logics       // 逻辑包，供api及service调用
│   ├── server.go    // 
│   └── service      // 服务层，包括动态扩容/自动权重/模型时效/consul静态ip/模型监控 等
└── util             // 一些有用的结构及方法
```

## 规范
1. logics的文件名尽量跟service/api中文件名保持一致

2. api路径跟路由定义保持一致。

3. 调用顺序

service -> logics -> dao 或 service -> dao

api -> logics -> dao 或 api -> dao

mysql等存储对象操作尽量放到dao中，不建议在service及api直接操作mysqlDB