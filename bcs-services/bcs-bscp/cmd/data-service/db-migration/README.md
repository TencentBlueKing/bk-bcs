
# db migration命令行工具使用


### 使用说明
- data-service的migrate子命令帮助信息
```bash
./data-service migrate -h
database migrations tool

Usage:
  bk-bscp-dataservice migrate [flags]
  bk-bscp-dataservice migrate [command]

Available Commands:
  create      create a new empty migrations file
  down        run down migrations
  status      display status of each migrations
  up          run up migrations

Flags:
  -h, --help   help for migrate

Global Flags:
  -b, --bind-ip ip                which IP the server is listen to
  -c, --config-file stringArray   the absolute path of the configuration file (repeatable)
  -v, --version                   show version

Use "bk-bscp-dataservice migrate [command] --help" for more information about a command.

```

- 创建migration 

```bash
# 创建一个migration，会在migrations目录下生成migration的go文件以及在migrations/sql目录下生成对应的sql文件
# 只需对生成的两个sql文件做修改
# 对于新加的migration，需要重新编译data-service服务，才能包含并可执行新的migration操作
$ ./data-service migrate create -n init_schema
Generated new migration files:
./db-migration/migrations/20230207215606_init_schema.go
./db-migration/migrations/sql/20230207215606_init_schema_up.sql
./db-migration/migrations/sql/20230207215606_init_schema_down.sql

# 为了便于演示，已经用上面同样的方式另外创建了两个测试用的migration，name参数分别为test_mig001和test_mig002
# 查看当前db的迁移状态，3个pending代表有三个migration待做迁移
$ ./data-service migrate status -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Migration 20230207165857_test_mig001 pending
Migration 20230207171029_test_mig002 pending
Migration 20230207215606_init_schema pending
```

- 向前迁移db
```bash
# 分步迁移，如向前迁移1个版本
$ ./data-service migrate up -s 1 -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Running migration 20230207165857
Finished running migration 20230207165857

# 直接迁移到最新版本
./data-service migrate up -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Running migration 20230207171029
Finished running migration 20230207171029
Running migration 20230207215606
Finished running migration 20230207215606

```

- 向后回滚db
```bash
# 分步回滚，如向后回滚1个版本
$ ./data-service migrate down -s 1 -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Reverting Migration 20230207215606
Finished reverting migration 20230207215606

# 直接回滚到最老版本
$ ./go-db-migration migrate down -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Reverting Migration 20230207171029
Finished reverting migration 20230207171029
Reverting Migration 20230207165857
Finished reverting migration 20230207165857
```

- 查看db迁移状态
```bash
# 打印当前db迁移状态，如下为迁移了一个版本时的db状态
# pending 代表有待执行的migration
# completed代表已执行过的migration
$ ./data-service migrate status -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Migration 20230207165857_test_mig001 completed
Migration 20230207171029_test_mig002 pending
Migration 20230207215606_init_schema pending


```

### 注意事项
- migrate的up、down、status子命令都需要连接mysql，所以需要用-c参数指定data-service的配置文件，用于获取mysql配置
- 对于新加的migration，需要重新编译data-service服务，才能包含并可执行新的migration操作
- data-service的migrate create命令需要在${bscp源码所在根目录}/cmd/data-service目录下运行，才能正常运行且保证生成的migration相关文件在正确位置
- migrate create指定的migration名称，中划线'-'会被转化成下划线'_'，以保持migration相关文件名称格式统一
