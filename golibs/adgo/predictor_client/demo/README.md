#使用
如果测试召回-广告侧向量模型调用demo，需要先修改build.sh 里面的build文件为 client_ads_retrieval_demo.go

#编译
./build.sh

#执行
./client_ads_retrieval_demo

#问题
如果报 "connect: connection refused", 说明服务端没有启动，可以联系算法工程同学开启
如果报 "rpc timeout", 建议适当调整超时时间或是做下重试处理