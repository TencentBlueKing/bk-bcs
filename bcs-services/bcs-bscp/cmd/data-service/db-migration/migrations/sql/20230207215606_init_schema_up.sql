/*
表结构说明：
各类模型表字段信息主要分为：
1. 主键id                        // 由 id_generators 表来管理当前模型的id最大值，用来实现主键id自增功能
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

# bk_bscp_admin is bscp system admin database, only save sharding_dbs, sharding_bizs, id_generators table.

create table if not exists `sharding_dbs`
(
    `id`         bigint(1) unsigned   not null auto_increment,

    # Spec is specifics of the resource defined with user
    `type`       varchar(20)          not null,
    `host`       varchar(64)          not null,
    `port`       smallint(1) unsigned not null,
    `user`       varchar(32)          not null,
    `password`   varchar(32)          not null,
    `database`   varchar(20)          not null,
    `memo`       varchar(256) default '',

    # Revision record revision info of the resource
    `creator`    varchar(64)          not null,
    `reviser`    varchar(64)          not null,
    `created_at` datetime(6)          not null,
    `updated_at` datetime(6)          not null,

    primary key (`id`),
    unique key `idx_host_port_user_passwd_db` (`host`, `port`, `user`, `password`, `database`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `sharding_bizs`
(
    `id`             bigint(1) unsigned not null auto_increment,

    # Spec is specifics of the resource defined with user
    `memo`           varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`         bigint(1) unsigned not null,
    `sharding_db_id` bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`        varchar(64)        not null,
    `reviser`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,
    `updated_at`     datetime(6)        not null,

    primary key (`id`),
    foreign key (`sharding_db_id`) references sharding_dbs (id),
    unique key `idx_bizID_dbID` (`biz_id`, `sharding_db_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `id_generators`
(
    `id`         bigint(1) unsigned not null auto_increment,
    `resource`   varchar(50)        not null,
    `max_id`     bigint(1) unsigned not null,
    `updated_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_resource` (`resource`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generators(resource, max_id, updated_at)
values ('applications', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('archived_apps', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('commits', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('config_items', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('contents', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('audits', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('events', 500, now());
insert into id_generators(resource, max_id, updated_at)
values ('current_released_instances', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('releases', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('released_config_items', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('strategies', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('current_published_strategies', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('published_strategy_histories', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('strategy_sets', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('resource_locks', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('groups', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('group_app_binds', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('released_groups', 0, now());

create table if not exists `archived_apps`
(
    `id`         bigint(1) unsigned not null,
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,
    `created_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `applications`
(
    `id`               bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`             varchar(255)       not null,
    `config_type`      varchar(20)        not null,
    `mode`             varchar(20)        not null,
    `memo`             varchar(256) default '',
    `reload_type`      varchar(20)  default '',
    `reload_file_path` varchar(255) default '',

    # Attachment is attachment info of the resource
    `biz_id`           bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`          varchar(64)        not null,
    `reviser`          varchar(64)        not null,
    `created_at`       datetime(6)        not null,
    `updated_at`       datetime(6)        not null,

    primary key (`id`),
    unique key `idx_bizID_name` (`biz_id`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `commits`
(
    `id`             bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `content_id`     bigint(1) unsigned not null,
    `signature`      varchar(64)        not null,
    `byte_size`      bigint(1) unsigned not null,
    `memo`           varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`         bigint(1) unsigned not null,
    `app_id`         bigint(1) unsigned not null,
    `config_item_id` bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,

    primary key (`id`),
    index `idx_bizID_appID_cfgID` (`biz_id`, `app_id`, `config_item_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `config_items`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`       varchar(255)       not null,
    `path`       varchar(255)       not null,
    `file_type`  varchar(20)        not null,
    `file_mode`  varchar(20)        not null,
    `memo`       varchar(256) default '',
    `user`       varchar(64)        not null,
    `user_group` varchar(64)        not null,
    `privilege`  varchar(64)        not null,

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`    varchar(64)        not null,
    `reviser`    varchar(64)        not null,
    `created_at` datetime(6)        not null,
    `updated_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_bizID_appID_name` (`biz_id`, `app_id`, `path`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `contents`
(
    `id`             bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `signature`      varchar(64)        not null,
    `byte_size`      bigint(1) unsigned not null,

    # Attachment is attachment info of the resource
    `biz_id`         bigint(1) unsigned not null,
    `app_id`         bigint(1) unsigned not null,
    `config_item_id` bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,

    primary key (`id`),
    index `idx_bizID_appID_cfgID` (`biz_id`, `app_id`, `config_item_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `audits`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `res_type`   varchar(50)        not null,
    `res_id`     bigint(1) unsigned not null,
    `action`     varchar(20)        not null,
    `rid`        varchar(64)        not null,
    `app_code`   varchar(64)        not null,
    `detail`     json               default null,

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned default null,

    `operator`   varchar(64)        not null,
    `created_at` datetime(6)        not null,

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

create table if not exists `events`
(
    `id`           bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `resource`     varchar(50)        not null,
    `op_type`      varchar(20)        not null,
    `resource_id`  bigint(1) unsigned default 0,
    `resource_uid` varchar(64)        default '',

    `final_status` tinyint unsigned   default 0,

    `biz_id`       bigint(1) unsigned not null,
    `app_id`       bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`      varchar(64)        not null,
    `created_at`   datetime(6)        not null,

    primary key (`id`),
    index `idx_resource_bizID` (`resource`, `biz_id`),
    index `idx_createdAt` (`created_at`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `current_released_instances`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `uid`        varchar(64)        not null,
    `release_id` bigint(1) unsigned not null,
    `memo`       varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`    varchar(64)        not null,
    `created_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_appID_uid` (`app_id`, `uid`),
    index `idx_uid` (`uid`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `releases`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`        varchar(255)       not null,
    `memo`        varchar(256)       default '',
    `deprecated`  boolean            not null,
    `publish_num` bigint(1) unsigned not null,

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`    varchar(64)        not null,
    `created_at` datetime(6)        not null,

    primary key (`id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `released_config_items`
(
    `id`             bigint(1) unsigned not null,

    `commit_id`      bigint(1) unsigned not null,
    `release_id`     bigint(1) unsigned not null,

    # Attachment is attachment info of the resource
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
    `memo`           varchar(256) default '',
    `user`           varchar(64)        not null,
    `user_group`     varchar(64)        not null,
    `privilege`      varchar(64)        not null,

    # Revision is reversion info of the resource being created.
    `creator`        varchar(64)        not null,
    `reviser`        varchar(64)        not null,
    `created_at`     datetime(6)        not null,
    `updated_at`     datetime(6)        not null,

    primary key (`id`),
    unique key `idx_releaseID_commitID` (`release_id`, `commit_id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `strategies`
(
    `id`              bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json         default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128) default '',
    `memo`            varchar(256) default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is attachment info of the resource
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`         varchar(64)        not null,
    `reviser`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,
    `updated_at`      datetime(6)        not null,

    primary key (`id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `groups`
(
    `id`                bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`              varchar(255)       not null,
    `mode`              varchar(20)        not null,
    `public`            boolean            default false,
    `selector`          json               default null,
    `uid`               varchar(64)        default '',

    # Attachment is attachment info of the resource
    `biz_id`            bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`           varchar(64)        not null,
    `reviser`           varchar(64)        not null,
    `created_at`        datetime(6)        not null,
    `updated_at`        datetime(6)        not null,

    primary key (`id`),
    unique key `idx_bizID_name` (`biz_id`, `name`),
    index `idx_bizID` (`biz_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `group_app_binds`
(
    `id`                bigint(1) unsigned not null,
    `group_id`          bigint(1) unsigned not null,
    `app_id`            bigint(1) unsigned not null,
    `biz_id`            bigint(1) unsigned not null,
    primary key (`id`),
    index `idx_groupID_appID_bizID` (`group_id`, `app_id`, `biz_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `released_groups`
(
    `id`                bigint(1) unsigned not null,
    `group_id`          bigint(1) unsigned not null,
    `app_id`            bigint(1) unsigned not null,
    `release_id`        bigint(1) unsigned not null,
    `strategy_id`       bigint(1) unsigned not null,
    `mode`              varchar(20)        not null,
    `selector`          json               default null,
    `uid`               varchar(64)        default '',
    `edited`            boolean            default false,
    `biz_id`            bigint(1) unsigned not null,
    `reviser`           varchar(64)        not null,
    `updated_at`        datetime(6)        not null,
    primary key (`id`),
    index `idx_groupID_appID_bizID` (`group_id`, `app_id`, `biz_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `current_published_strategies`
(
    `id`              bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json         default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128) default '',
    `memo`            varchar(256) default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is attachment info of the resource
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,
    `strategy_id`     bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,

    primary key (`id`),
    unique key `idx_strategyID` (`strategy_id`),
    index `idx_appID` (`app_id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`),
    index `idx_bizID_releaseID` (`biz_id`, `release_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `published_strategy_histories`
(
    `id`              bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`            varchar(255)       not null,
    `release_id`      bigint(1) unsigned not null,
    `as_default`      boolean            not null,
    `scope`           json         default null,
    `mode`            varchar(20)        not null,
    `namespace`       varchar(128) default '',
    `memo`            varchar(256) default '',
    `pub_state`       varchar(20)        not null,

    # Attachment is attachment info of the resource
    `biz_id`          bigint(1) unsigned not null,
    `app_id`          bigint(1) unsigned not null,
    `strategy_set_id` bigint(1) unsigned not null,
    `strategy_id`     bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`         varchar(64)        not null,
    `created_at`      datetime(6)        not null,

    primary key (`id`),
    index `idx_bizID_appID_setID_strategyID` (`biz_id`, `app_id`, `strategy_set_id`, `strategy_id`),
    index `idx_bizID_appID_setID_namespace` (`biz_id`, `app_id`, `strategy_set_id`, `namespace`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `strategy_sets`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`       varchar(255)       not null,
    `mode`       varchar(20)        not null,
    `status`     varchar(20)  default '',
    `memo`       varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,
    `app_id`     bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`    varchar(64)        not null,
    `reviser`    varchar(64)        not null,
    `created_at` datetime(6)        not null,
    `updated_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_appID_name` (`app_id`, `name`),
    unique key `idx_bizID_id` (`biz_id`, `id`),
    index `idx_bizID_appID` (`biz_id`, `app_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `resource_locks`
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
