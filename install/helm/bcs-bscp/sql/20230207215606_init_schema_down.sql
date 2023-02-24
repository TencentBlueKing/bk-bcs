/*
删除初始化migration所创建的所有表
删除顺序与创建顺序相反，避免出现类似' referenced by a foreign key constraint '的报错
*/

drop table if exists `resource_lock`;
drop table if exists `strategy_set`;
drop table if exists `published_strategy_history`;
drop table if exists `current_published_strategy`;
drop table if exists `strategy`;
drop table if exists `released_config_item`;
drop table if exists `releases`;
drop table if exists `current_released_instance`;
drop table if exists `event`;
drop table if exists `audit`;
drop table if exists `content`;
drop table if exists `config_item`;
drop table if exists `commits`;
drop table if exists `application`;
drop table if exists `archived_app`;
drop table if exists `id_generator`;
drop table if exists `sharding_biz`;
drop table if exists `sharding_db`;