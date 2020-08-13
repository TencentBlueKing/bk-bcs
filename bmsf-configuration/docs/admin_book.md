# 系统管理命令

系统管理命令仅限管理员使用，包含数据库信息映射，业务创建，业务数据存储关联。`相关的操作命令默认不会在--help中显示出来`。

## Bussiness 操作

**初始化业务**，初始化业务时，如果确认需要存储隔离，为业务分配单独的数据库，需要提前在mysql中创建数据库。

```shell
> bk-bscp-client create business --help
Create new business, only affected in administrator mode

Usage:
  bk-bscp-client create business [flags]

Aliases:
  business, bus

Examples:

  bscp-client create business --file newbusiness.yaml

  yaml format template as followed:
  kind: bscp-business
  version: 0.1.1
  spec:
	name: X-Game
	deptID: bking
	creator: MrMGXXXX
	auth: ${user}:${pwd}
	memo: annotation
  db:
	#dbID sharding index,对应mysql instance
	dbID: bscp-default-sharding
	#对应mysql中不同数据库
	dbName: bscp-default
	#如果有以下信息说明是新建sharedingDB
	host: 127.0.0.1
	port: 3306
	user: mysql
	password: ${pwd}
	memo: information


Flags:
  -f, --file string   settings new business yaml file
  -h, --help          help for business

Global Flags:
      --business string   business Name to operate. Get parameter priority: command -> env -> .bscp/desc
      --operator string   user name for operation.  Get parameter priority: command -> env -> .bscp/desc
      --token string      user token for operation. Get parameter priority: command -> env -> .bscp/desc
```

**查看业务列表**

```shell
> bk-bscp-client list business
+----------------------------------------+---------+------------+---------+-------+
|                   ID                   |  NAME   | DEPARTMENT | CREATOR | STATE |
+----------------------------------------+---------+------------+---------+-------+
| B-48de67cb-b6d5-11ea-90b2-5254006865b1 | X-Game  | bking      | guohu   |     0 |
+----------------------------------------+---------+------------+---------+-------+
```

## ShardingDB 操作

**创建数据库实例**

```shell
> bk-bscp-client create shardingdb --dbid bscp-test-shardingdb --host 127.0.1.1 --port 3307 --user guohu --password admin --memo "this is a example for create"
create shardingDB successfully:  bscp-test-shardingdb
```

**更新数据库实例**

```shell
> bk-bscp-client update shardingdb --dbid bscp-test-shardingdb --host 127.0.0.1 --port 3306 --user guohu --password admin --memo "this is a example for update"
Update resources successfully
```

**查看数据库实例**

```shell
> bk-bscp-client list shardingdb
+----------------------+-----------+------+------+-------+---------------------+
|          ID          |   HOST    | PORT | USER | STATE |     LASTUPDATED     |
+----------------------+-----------+------+------+-------+---------------------+
| bscp-test-shardingdb | 127.0.0.1 | 3306 | guohu |     0 | 2020-07-23 23:31:41 |
+----------------------+-----------+------+-------+-------+---------------------+

> bk-bscp-client get shardingdb --dbid bscp-test-shardingdb
DBID: 		bscp-test-shardingdb
Host: 		127.0.0.1
Port: 		3306
User: 		guohu
Password: 	admin
Memo: 		this is a example for update
State:		Affectived
CreatedAt: 	2020-07-23 23:30:13
UpdatedAt: 	2020-07-23 23:31:41
```

### Sharding 操作

**创建 Sharding**

```shell
> bk-bscp-client create sharding --key B-b02ed6f1-cc2f-11ea-9cfe-5254006865b1 --dbid bscp-lolo-shardingdb --dbname testDB --memo "this is a create example"
create sharding successfully: B-b02ed6f1-cc2f-11ea-9cfe-5254006865b1
```

**更新 Sharding**

```shell
> bk-bscp-client update sharding --key B-b02ed6f1-cc2f-11ea-9cfe-5254006865b1 --dbid bscp-lolo-shardingdb --dbname bscplolo --memo "this is a update example"
Update resources successfully
```

**查看 Sharding**

```shell
> bk-bscp-client get sharding --key B-b02ed6f1-cc2f-11ea-9cfe-5254006865b1
Key: 		B-b02ed6f1-cc2f-11ea-9cfe-5254006865b1
DBID: 		bscp-lolo-shardingdb
DBName: 	bscplolo
Memo: 		this is a update example
State:		Affectived
CreatedAt: 	2020-07-23 23:35:27
UpdatedAt: 	2020-07-23 23:38:01
```
