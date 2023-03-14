/*
表结构说明：
各类模型表字段信息主要分为：
1. 主键id                        // 由 id_generator 表来管理当前模型的id最大值，用来实现主键id自增功能
3. 模型特定字段信息Spec            // 需要用户特殊定义的字段 (Spec)
2. 外键id                        // 和当前模型有关联关系的模型主键id (Attachment)
4. 关联模型特定字段信息Spec        // 当前模型需要记录有关联关系的模型特定字段信息 (OtherSpec)
5. 创建信息（CreatedRevision）、创建及修正信息（Revision）

注:
    1. 字段说明统一参照 pkg/dal/table 目录下的数据结构定义说明。
    2. varchar字符类型实际存储大小为 Len + 存储长度大小(1/2字节)，但是索引是根据设定的varchar长度进行建立的，
    如需要对字段建立索引，注意存储消耗。varchar类型字段长度从小于255范围，扩展到大于255范围，因为记录varchar
    实际长度的字符需要从 1byte -> 2byte，会进行锁表。所以，表字段跨255范围扩展，需确认影响。
    3. 各类表的name以及namespace字段采用varchar第一范围最大值(255)进行存储，memo字段采用第二范围最小值(256)进行存储，
    非必要禁止跨界。

Sql语句规范：
1. 字段需要按照上述分类进行排序和分类。
*/

# bk_bscp_admin is bscp system admin database, only save sharding_db、sharding_biz、id_generator table.
create database if not exists bk_bscp_admin;
use bk_bscp_admin;

create table if not exists `sharding_db`
(
    `id`         bigint(1) unsigned   not null auto_increment,

    # Spec is a collection of resource's specifics defined with user
    `type`       varchar(20)          not null,
    `host`       varchar(64)          not null,
    `port`       smallint(1) unsigned not null,
    `user`       varchar(32)          not null,
    `password`   varchar(32)          not null,
    `database`   varchar(20)          not null,
    `memo`       varchar(256)       default '',

    # Revision record this resource's revision information
    `creator`    varchar(64)          not null,
    `reviser`    varchar(64)          not null,
    `created_at` datetime(6)          not null,
    `updated_at` datetime(6)          not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_host_port_user_passwd_db` (`host`, `port`, `user`, `password`, `database`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `sharding_biz`
(
    `id`             bigint(1) unsigned not null auto_increment,

    # Spec is a collection of resource's specifics defined with user
    `memo`           varchar(256)       default '',

    # Attachment is a resource attachment id
    `biz_id`         bigint(1) unsigned not null,
    `sharding_db_id` bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`        varchar(64)        not null,
    `reviser`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,
    `updated_at`     datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`      varchar(255)       default '',
    `reservedB`      varchar(255)       default '',
    `reservedC`      bigint(1) unsigned default 0,

    primary key (`id`),
    foreign key (`sharding_db_id`) references sharding_db (id),
    unique key `idx_bizID_dbID` (`biz_id`, `sharding_db_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `id_generator`
(
    `id`         bigint(1) unsigned not null auto_increment,
    `resource`   varchar(50)        not null,
    `max_id`     bigint(1) unsigned not null,
    `updated_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_resource` (`resource`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(id, resource, max_id, updated_at)
values (1, 'application', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (2, 'archived_app', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (3, 'commits', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (4, 'config_item', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (5, 'content', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (6, 'audit', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (7, 'event', 500, now());
insert into id_generator(id, resource, max_id, updated_at)
values (8, 'current_released_instance', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (9, 'releases', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (10, 'released_config_item', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (11, 'strategy', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (12, 'current_published_strategy', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (13, 'published_strategy_history', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (14, 'strategy_set', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (15, 'resource_lock', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (16, 'group_', 0, now());
insert into id_generator(id, resource, max_id, updated_at)
values (17, 'group_category', 0, now());

create table if not exists `archived_app`
(
    `id`         bigint(1) unsigned not null,
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,
    `created_at` datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `application`
(
    `id`               bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`             varchar(255)       not null,
    `config_type`      varchar(20)        not null,
    `mode`             varchar(20)        not null,
    `memo`             varchar(256)       default '',
    `reload_type`      varchar(20)        default '',
    `reload_file_path` varchar(255)       default '',

    # Attachment is a resource attachment id
    `biz_id`           bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`          varchar(64)        not null,
    `reviser`          varchar(64)        not null,
    `created_at`       datetime(6)        not null,
    `updated_at`       datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`        varchar(255)       default '',
    `reservedB`        varchar(255)       default '',
    `reservedC`        bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_bizID_name` (`biz_id`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `commits`
(
    `id`             bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `content_id`     bigint(1) unsigned not null,
    `signature`      varchar(64)        not null,
    `byte_size`      bigint(1) unsigned not null,
    `memo`           varchar(256)       default '',

    # Attachment is a resource attachment id
    `biz_id`         bigint(1) unsigned not null,
    `app_id`         bigint(1) unsigned not null,
    `config_item_id` bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`      varchar(255)       default '',
    `reservedB`      varchar(255)       default '',
    `reservedC`      bigint(1) unsigned default 0,

    primary key (`id`),
    index `idx_bizID_appID_cfgID` (`biz_id`, `app_id`, `config_item_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `config_item`
(
    `id`         bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`       varchar(255)       not null,
    `path`       varchar(255)       not null,
    `file_type`  varchar(20)        not null,
    `file_mode`  varchar(20)        not null,
    `memo`       varchar(256)       default '',
    `user`       varchar(64)        not null,
    `user_group` varchar(64)        not null,
    `privilege`  varchar(64)        not null,

    # Attachment is a resource attachment id
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`    varchar(64)        not null,
    `reviser`    varchar(64)        not null,
    `created_at` datetime(6)        not null,
    `updated_at` datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_bizID_appID_name` (`biz_id`, `app_id`, `path`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `content`
(
    `id`             bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `signature`      varchar(64)        not null,
    `byte_size`      bigint(1) unsigned not null,

    # Attachment is a resource attachment id
    `biz_id`         bigint(1) unsigned not null,
    `app_id`         bigint(1) unsigned not null,
    `config_item_id` bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`      varchar(255)       default '',
    `reservedB`      varchar(255)       default '',
    `reservedC`      bigint(1) unsigned default 0,

    primary key (`id`),
    index `idx_bizID_appID_cfgID` (`biz_id`, `app_id`, `config_item_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `audit`
(
    `id`         bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `res_type`   varchar(50)        not null,
    `res_id`     bigint(1) unsigned not null,
    `action`     varchar(20)        not null,
    `rid`        varchar(64)        not null,
    `app_code`   varchar(64)        not null,
    `detail`     json               default null,

    # Attachment is a resource attachment id
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned default null,

    `operator`   varchar(64)        not null,
    `created_at` datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    key `idx_resType` (`res_type`),
    key `idx_resID` (`res_id`),
    key `idx_action` (`action`),
    key `idx_createdAt` (`created_at`),
    key `idx_bizID_appID_createdAt` (`biz_id`, `app_id`, `created_at`),
    key `idx_bizID_resType_createdAt` (`biz_id`, `res_type`, `created_at`),
    key `idx_bizID_operator_createdAt` (`biz_id`, `operator`, `created_at`),
    key `idx_appCode_operator_createdAt` (`app_code`, `operator`, `created_at`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `event`
(
    `id`           bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `resource`     varchar(50)        not null,
    `op_type`      varchar(20)        not null,
    `resource_id`  bigint(1) unsigned default 0,
    `resource_uid` varchar(64)        default '',

    `final_status` tinyint unsigned   default 0,

    `biz_id`       bigint(1) unsigned not null,
    `app_id`       bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`      varchar(64)        not null,
    `created_at`   datetime(6)        not null,

    primary key (`id`),
    index `idx_resource_bizID` (`resource`, `biz_id`),
    index `idx_createdAt` (`created_at`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `current_released_instance`
(
    `id`         bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `uid`        varchar(64)        not null,
    `release_id` bigint(1) unsigned not null,
    `memo`       varchar(256)       default '',

    # Attachment is a resource attachment id
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`    varchar(64)        not null,
    `created_at` datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_appID_uid` (`app_id`, `uid`),
    index `idx_uid` (`uid`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `releases`
(
    `id`         bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`        varchar(255)       not null,
    `memo`        varchar(256)       default '',
    `deprecated`  boolean            not null,
    `publish_num` bigint(1) unsigned not null,
 
    # Attachment  is a resource attachment id
    `biz_id`      bigint(1) unsigned not null,
    `app_id`      bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`     varchar(64)        not null,
    `created_at`  datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`   varchar(255)       default '',
    `reservedB`   varchar(255)       default '',
    `reservedC`   bigint(1) unsigned default 0,

    primary key (`id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `released_config_item`
(
    `id`             bigint(1) unsigned not null,

    `commit_id`      bigint(1) unsigned not null,
    `release_id`     bigint(1) unsigned not null,

    # Attachment is a resource attachment id
    `biz_id`         bigint(1) unsigned not null,
    `app_id`         bigint(1) unsigned not null,
    `config_item_id` bigint(1) unsigned not null,

    # Commit Spec
    `content_id`     bigint(1) unsigned not null,
    `signature`      varchar(64)        not null,
    `byte_size`      bigint(1) unsigned not null,

    # Config Item Spec
    `name`           varchar(255)       not null,
    `path`           varchar(255)       not null,
    `file_type`      varchar(20)        not null,
    `file_mode`      varchar(20)        not null,
    `memo`           varchar(256)       default '',
    `user`           varchar(64)        not null,
    `user_group`     varchar(64)        not null,
    `privilege`      varchar(64)        not null,

    # Revision record this resource's revision information
    `creator`        varchar(64)        not null,
    `reviser`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,
    `updated_at`     datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`      varchar(255)       default '',
    `reservedB`      varchar(255)       default '',
    `reservedC`      bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_releaseID_commitID` (`release_id`, `commit_id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `strategy`
(
    `id`              bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json               default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128)       default '',
    `memo`            varchar(256)       default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is a resource attachment id
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`         varchar(64)        not null,
    `reviser`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,
    `updated_at`      datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`       varchar(255)       default '',
    `reservedB`       varchar(255)       default '',
    `reservedC`       bigint(1) unsigned default 0,

    primary key (`id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `group_`
(
    `id`                bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`              varchar(255)       not null,
    `mode`              varchar(20)        not null,
    `selector`          json               default null,
    `uid`               varchar(64)        default '',

    # Attachment is a resource attachment id
    `biz_id`            bigint(1) unsigned not null,
    `app_id`            bigint(1) unsigned not null,
    `group_category_id` bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`           varchar(64)        not null,
    `reviser`           varchar(64)        not null,
    `created_at`        datetime(6)        not null,
    `updated_at`        datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`         varchar(255)       default '',
    `reservedB`         varchar(255)       default '',
    `reservedC`         bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_categoryID_name` (`group_category_id`, `name`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `group_category`
(
    `id`              bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`            varchar(255)       not null,

    # Attachment is a resource attachment id
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`       varchar(255)       default '',
    `reservedB`       varchar(255)       default '',
    `reservedC`       bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_bizID_appID_name` (`biz_id`, `app_id`, `name`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `current_published_strategy`
(
    `id`              bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json               default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128)       default '',
    `memo`            varchar(256)       default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is a resource attachment id
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,
    `strategy_id`     bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`       varchar(255)       default '',
    `reservedB`       varchar(255)       default '',
    `reservedC`       bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_strategyID` (`strategy_id`),
    index `idx_appID` (`app_id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`),
    index `idx_bizID_releaseID` (`biz_id`, `release_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `published_strategy_history`
(
    `id`              bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json               default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128)       default '',
    `memo`            varchar(256)       default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is a resource attachment id
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,
    `strategy_id`     bigint(1) unsigned not null,

    # CreatedRevision is a resource's reversion information being created.
    `creator`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`       varchar(255)       default '',
    `reservedB`       varchar(255)       default '',
    `reservedC`       bigint(1) unsigned default 0,

    primary key (`id`),
    index `idx_bizID_appID_setID_strategyID` (`biz_id`, `app_id`, `strategy_set_id`, `strategy_id`),
    index `idx_bizID_appID_setID_namespace` (`biz_id`, `app_id`, `strategy_set_id`, `namespace`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `strategy_set`
(
    `id`         bigint(1) unsigned not null,

    # Spec is a collection of resource's specifics defined with user
    `name`       varchar(255)       not null,
    `mode`       varchar(20)        not null,
    `status`     varchar(20)        default '',
    `memo`       varchar(256)       default '',

    # Attachment is a resource attachment id
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # Revision record this resource's revision information
    `creator`    varchar(64)        not null,
    `reviser`    varchar(64)        not null,
    `created_at` datetime(6)        not null,
    `updated_at` datetime(6)        not null,

    # Reserve reserve field.
    `reservedA`  varchar(255)       default '',
    `reservedB`  varchar(255)       default '',
    `reservedC`  bigint(1) unsigned default 0,

    primary key (`id`),
    unique key `idx_appID_name` (`app_id`, `name`),
    unique key `idx_bizID_id` (`biz_id`, `id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `resource_lock`
(
    `id`        bigint(1) unsigned not null auto_increment,
    `res_type`  varchar(50)        not null,
    `res_key`   varchar(512)       not null,
    `res_count` bigint(1) unsigned not null,

    `biz_id`    bigint(1) unsigned not null,

    primary key (`id`),
    unique key `idx_bizID_resType_resKey` (`biz_id`, `res_type`, `res_key`)
) engine = innodb
  default charset = utf8mb4;
