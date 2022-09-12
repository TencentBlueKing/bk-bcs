# bcs-helm-manager-migrator

从 bcs-ui 同步 helm 数据到 bcs-helm-manager 模块中。

目前同步的数据：
- chart 仓库信息(项目仓库地址和账号密码)

## bcs-ui helm 数据库结构

helm_repository
```
+----------------+--------------+------+-----+---------+----------------+
| Field          | Type         | Null | Key | Default | Extra          |
+----------------+--------------+------+-----+---------+----------------+
| id             | int(11)      | NO   | PRI | NULL    | auto_increment |
| created_at     | datetime     | NO   |     | NULL    |                |
| updated_at     | datetime     | NO   |     | NULL    |                |
| url            | varchar(200) | NO   |     | NULL    |                |
| name           | varchar(32)  | NO   |     | NULL    |                |
| description    | varchar(512) | NO   |     | NULL    |                |
| project_id     | varchar(32)  | NO   | MUL | NULL    |                |
| provider       | varchar(32)  | NO   |     | NULL    |                |
| is_provisioned | tinyint(1)   | NO   |     | NULL    |                |
| refreshed_at   | datetime     | YES  |     | NULL    |                |
| commit         | varchar(64)  | YES  |     | NULL    |                |
| branch         | varchar(30)  | YES  |     | NULL    |                |
| storage_info   | longtext     | NO   |     | NULL    |                |
+----------------+--------------+------+-----+---------+----------------+
```

helm_repo_auth
```
+-------------+-------------+------+-----+---------+----------------+
| Field       | Type        | Null | Key | Default | Extra          |
+-------------+-------------+------+-----+---------+----------------+
| id          | int(11)     | NO   | PRI | NULL    | auto_increment |
| type        | varchar(16) | NO   |     | NULL    |                |
| credentials | longtext    | NO   |     | NULL    |                |
| role        | varchar(16) | NO   |     | NULL    |                |
| repo_id     | int(11)     | NO   | MUL | NULL    |                |
+-------------+-------------+------+-----+---------+----------------+
```