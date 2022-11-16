## 运行集成测试相关说明

### 1. 文件说明
```shell
.
├── application.test
├── *.test
├── start.sh
└── tools.sh
```

- *.test 是编译完的二进制测试文件，文件名代表是对什么类型资源进行测试。如 application.test 是对 application 资源相关接口的集成测试。
- tools.sh 是对测试执行结果导出的 json 文件进行统计分析的工具，最终会生成一个 html 页面。
- start.sh 执行测试并进行统计分析的脚本，需要配置相关环境变量才可运行。

#### 1.1 start.sh 执行需设置的环境变量
```shell
# 集成测试环境 api-server 地址
export ENV_SUITE_TEST_REQUEST_HOST=http://127.0.0.1:8080

# mysql 相关配置， 运行集成测试之前会进行清库操作，否则会对测试结果造成影响
export ENV_SUITE_TEST_MYSQL_IP=127.0.0.1
export ENV_SUITE_TEST_MYSQL_PORT=3306
export ENV_SUITE_TEST_MYSQL_USER=root
export ENV_SUITE_TEST_MYSQL_PW=admin
export ENV_SUITE_TEST_MYSQL_DB=bk_bscp_admin

# 测试结果导出的json文件存储目录
export ENV_SUITE_TEST_SAVE_DIR=./result
# 对测试结果统计分析生成的html页面存储路径
export ENV_SUITE_TEST_OUTPUT_PATH=./result/statistics.html
```