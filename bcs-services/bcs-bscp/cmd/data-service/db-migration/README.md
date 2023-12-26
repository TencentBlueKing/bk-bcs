# db migration命令行工具使用

### 使用说明

- data-service的migrate子命令帮助信息

```bash
./bk-bscp-dataservice migrate -h
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
  -d, --debug    whether debug gorm to print sql, default is false
  -h, --help     help for migrate

Global Flags:
  -b, --bind-ip ip                which IP the server is listen to
  -c, --config-file stringArray   the absolute path of the configuration file (repeatable)
  -v, --version                   show version

Use "bk-bscp-dataservice migrate [command] --help" for more information about a command.
```

- 创建migration

**支持两种db迁移模式：sql和gorm**

**gorm模式（默认模式，创建时的mode参数为gorm）**
```bash
# gorm模式下创建一个migration，会在migrations目录下生成migration的go文件
# 只需对生成的该go文件做修改
# 对于新加的migration，需要重新编译data-service服务，才能包含并可执行新的migration操作
# 通过命令行参数指定gorm模式：-m gorm 或 --mode gorm (也可不指定，默认为该模式)
$ ./bk-bscp-dataservice migrate create -n init_schema
Generated new migration files:
./cmd/data-service/db-migration/migrations/20230520120159_init_schema.go

# 为了便于演示，已经用上面同样的方式另外创建了两个测试用的migration，name参数分别为test_mig001和test_mig002
# 查看当前db的迁移状态，3个pending代表有三个migration待做迁移
$ ./bk-bscp-dataservice migrate status -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Migration 20230207165857_test_mig001 pending
Migration 20230207171029_test_mig002 pending
Migration 20230207215606_init_schema pending
```

**sql模式（创建时的mode参数为sql）**
```bash
# sql模式下创建一个migration，会在migrations目录下生成migration的go文件以及在migrations/sql目录下生成对应的sql文件
# 只需对生成的两个sql文件做修改
# 对于新加的migration，需要重新编译data-service服务，才能包含并可执行新的migration操作
# 通过命令行参数指定sql模式：-m sql 或 --mode sql
$ ./bk-bscp-dataservice migrate create -n init_schema -m sql
Generated new migration files:
./cmd/data-service/db-migration/migrations/20230207215606_init_schema.go
./cmd/data-service/db-migration/migrations/sql/20230207215606_init_schema_up.sql
./cmd/data-service/db-migration/migrations/sql/20230207215606_init_schema_down.sql

# 为了便于演示，已经用上面同样的方式另外创建了两个测试用的migration，name参数分别为test_mig001和test_mig002
# 查看当前db的迁移状态，3个pending代表有三个migration待做迁移
$ ./bk-bscp-dataservice migrate status -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Migration 20230207165857_test_mig001 pending
Migration 20230207171029_test_mig002 pending
Migration 20230207215606_init_schema pending
```

- 向前迁移db

```bash
# 分步迁移，如向前迁移1个版本
$ ./bk-bscp-dataservice migrate up -s 1 -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Running migration 20230207165857
Finished running migration 20230207165857

# 直接迁移到最新版本
./bk-bscp-dataservice migrate up -c /tmp/data_service.yaml
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
$ ./bk-bscp-dataservice migrate down -s 1 -c /tmp/data_service.yaml
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
$ ./bk-bscp-dataservice migrate status -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Migration 20230207165857_test_mig001 completed
Migration 20230207171029_test_mig002 pending
Migration 20230207215606_init_schema pending

```

- gorm模式下开启debug日志

```bash
# 为了便于排查gorm模式下db迁移过程中的错误，可开启debug日志，打印执行的sql语句
# 通过命令行参数指定gorm的debug日志：-d 或 --debug
$ ./bk-bscp-dataservice migrate up -s 1 -d -c /tmp/data_service.yaml
Connecting to MySQL database...
Database connected!
Running migration 20230511114513

2023/05/20 12:09:50 github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrations/20230511114513_add_template.go:121
[1.358ms] [rows:-] SELECT DATABASE()

2023/05/20 12:09:50 github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrations/20230511114513_add_template.go:121
[4.871ms] [rows:1] SELECT SCHEMA_NAME from Information_schema.SCHEMATA where SCHEMA_NAME LIKE 'bk_bscp_admin%' ORDER BY SCHEMA_NAME='bk_bscp_admin' DESC,SCHEMA_NAME limit 1

2023/05/20 12:09:50 github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrations/20230511114513_add_template.go:121
[4.914ms] [rows:-] SELECT count(*) FROM information_schema.tables WHERE table_schema = 'bk_bscp_admin' AND table_name = 'template_spaces' AND table_type = 'BASE TABLE'

2023/05/20 12:09:50 github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrations/20230511114513_add_template.go:121
[48.765ms] [rows:0] CREATE TABLE `template_spaces` (`id` bigint(1) unsigned not null,`name` varchar(255) not null,`memo` varchar(256) default '',`biz_id` bigint(1) unsigned not null,`creator` varchar(64) not null,`reviser` varchar(64) not null,`created_at` datetime(6) not null,`updated_at` datetime(6) not null,PRIMARY KEY (`id`),UNIQUE INDEX `idx_bizID_name` (`biz_id`,`name`))ENGINE=InnoDB CHARSET=utf8mb4
```


### 注意事项

- migrate的up、down、status子命令都需要连接mysql，所以需要用-c参数指定data-service的配置文件，用于获取mysql配置
- 对于新加的migration，需要重新编译data-service服务，才能包含并可执行新的migration操作
- data-service的migrate create命令需要在bscp源码根目录下运行，才能正常运行且保证生成的migration相关文件在正确位置
- migrate create指定的migration名称，中划线'-'会被转化成下划线'_'，以保持migration相关文件名称格式统一