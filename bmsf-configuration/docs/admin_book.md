# 系统管理命令

系统管理命令仅限管理员使用，包含数据库信息映射，业务创建，业务数据存储关联。`相关的操作命令默认不会在--help中显示出来`。

**初始化业务**，初始化业务时，如果确认需要存储隔离，为业务分配单独的数据库，需要提前在mysql中创建数据库。

```shell
> bk-bscp-client create business --help
create new business, only affected in administrator mode

Usage:
  bscp-client create business [flags]

Aliases:
  business, bu, bi

Examples:

bscp-client create business --file newbusiness.yaml
bscp-client create business -f business.yaml

yaml format template as followed:
    kind: bscp-business
    version: 0.1.1
    spec:
        name: X-Game
        deptID: bking
        creator: melo
        memo: annotation
    db:
        # 数据库实例, 对应MySQL sharding db instance ID
        dbID: bscp-main-shardingdb

        # 对应MySQL中不同database
        dbName: bscp_xgame

        # 如果有以下信息说明是新建shareding db
        host: 127.0.0.1
        port: 3306
        user: mysql
        password: ${pwd}
        memo: information


Flags:
  -f, --file string   settings new business yaml file
  -h, --help          help for business

Global Flags:
      --business string     Business Name to operate. Also comes from ENV BSCP_BUSINESS
      --configfile string   BlueKing Service Configuration Platform CLI configuration. (default "/etc/bscp/client.yaml")
      --operator string     user name for operation, use for audit, Also comes from ENV BSCP_OPERATOR
```

**查看业务列表**

```shell
> bk-bscp-client list business
+----------------------------------------+------------+------------+---------+-------+
|                    ID                  |    NAME    | DEPARTMENT | CREATOR | STATE |
+----------------------------------------+------------+------------+---------+-------+
| B-b9f8492c-cedf-11e9-9e47-5254000ea971 |  X-Game    |   bking    |   melo  |     0 |
+----------------------------------------+------------+------------+---------+-------+
```

**查看数据库实例**

```shell
> bk-bscp-client list db
+-----------------------+--------------+------+------+-------+---------------------+
|          ID           |     HOST     | PORT | USER | STATE |     LASTUPDATED     |
+-----------------------+--------------+------+------+-------+---------------------+
|  bscp-main-shardingdb |  127.0.0.1   | 3306 | root |     0 | 2019-09-04 14:44:46 |
+-----------------------+--------------+------+------+-------+---------------------+
```
