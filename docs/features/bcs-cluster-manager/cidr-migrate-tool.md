# cidr数据迁移工具

## 将mysql中的数据下载到本地文件

```shell
./cidr-migration-tool --work-mode dumps \
    --dsn "xxxxxx:xxxxxxx@tcp(mysql.addr:3306)/databasename?charset=utf8mb4&parseTime=True&loc=Local" \
    --filename cidr.json
```

## 将本地文件上传至mongo

```shell
./cidrmigrate --mode upload \
    --filename cidr.json \
    --mongo_address "127.0.0.1:27018" \
    --mongo_database xxxxxxxxx \
    --mongo_username xxxxxxxxx \
    --mongo_password xxxxxxxxx
```
