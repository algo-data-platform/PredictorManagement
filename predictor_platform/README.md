本目录下主要有三个子目录：
1) frontend: 为前度数据展示，采用vue+javascript，来渲染展示server模块采集的数据。
如需修改前端显示及逻辑进入frontend/src目录编辑相关文件
前端编译：进入server目录执行sh build.sh在server目录会生成frontend目录
前端编译依赖：axios（npm install axios）

2）server: 该服务主要为后端数据处理部分。
执行主程序前必须编译前端代码如上所述
线上：go run main.go或者go build main.go, 然后直接运行./main
线下调试：需要指定config文件：go run main.go --conf=./runtime/config.json 或者 go build main.go, 然后直接运行./main --conf=./runtime/config.json
运行主程序后访问前端url: http://ip:port  其中ip和port为conf文件里定义的http_host和http_port

3) deploy：该目录主要有线上部署脚本及依赖文件
