## 运行压力测试相关说明

### 1. 文件说明

```shell
.
├── cache_service.test
├── feed_server_grpc.test
├── feed_server_http.test
├── start.sh
└── tools
    └── gen-data
```

- *.test 是编译完的二进制压测文件，文件名代表是对什么服务进行压测。如 cache_service.test 是对 cacheservice 相关接口的压测。
- tools/gen-data 是用于生成压测数据的脚本，生成数据需要调整db隔离级别为读已提交，脚本为并发执行，如果不调整，会死锁，如果需要导出数据，将 audit 清空。
- start.sh 执行压测并生成压测报告。

#### 1.1 start.sh 执行需设置的环境变量

```shell
# 压测环境 cache-service/feed-server 地址
export ENV_BENCH_TEST_CACHE_REQUEST_HOST=127.0.0.1:9514
export ENV_BENCH_TEST_FEED_REQUEST_HOST=http://127.0.0.1:9610

# mysql 相关配置， 运行集成测试之前会进行清库操作，否则会对测试结果造成影响
export ENV_SUITE_TEST_MYSQL_IP=127.0.0.1
export ENV_SUITE_TEST_MYSQL_PORT=3306
export ENV_SUITE_TEST_MYSQL_USER=root
export ENV_SUITE_TEST_MYSQL_PW=admin
export ENV_SUITE_TEST_MYSQL_DB=bk_bscp_admin

# 对压测结果统计分析生成的html页面存储路径
export ENV_SUITE_TEST_OUTPUT_PATH=./bench_report.html
```
