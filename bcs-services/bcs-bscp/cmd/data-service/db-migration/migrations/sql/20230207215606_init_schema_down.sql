/*
删除初始化migration所创建的所有表
删除顺序与表依赖顺序相反，避免出现类似' referenced by a foreign key constraint '的报错
*/

drop table if exists `resource_locks`;
drop table if exists `strategy_sets`;
drop table if exists `published_strategy_histories`;
drop table if exists `current_published_strategies`;
drop table if exists `groups`;
drop table if exists `group_app_binds`;
drop table if exists `released_groups`;
drop table if exists `released_config_items`;
drop table if exists `releases`;
drop table if exists `current_released_instances`;
drop table if exists `events`;
drop table if exists `audits`;
drop table if exists `contents`;
drop table if exists `config_items`;
drop table if exists `commits`;
drop table if exists `applications`;
drop table if exists `archived_apps`;
drop table if exists `strategies`;
drop table if exists `id_generators`;
drop table if exists `sharding_bizs`;
drop table if exists `sharding_dbs`;
