create table if not exists `template_spaces`
(
    `id`         bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`       varchar(255)       not null,
    `memo`       varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`     bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`    varchar(64)        not null,
    `reviser`    varchar(64)        not null,
    `created_at` datetime(6)        not null,
    `updated_at` datetime(6)        not null,

    primary key (`id`),
    unique key `idx_bizID_name` (`biz_id`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `templates`
(
    `id`                bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`              varchar(255)       not null,
    `path`              varchar(255)       not null,
    `memo`              varchar(256) default '',

    # Attachment is attachment info of the resource
    `biz_id`            bigint(1) unsigned not null,
    `template_space_id` bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`           varchar(64)        not null,
    `reviser`           varchar(64)        not null,
    `created_at`        datetime(6)        not null,
    `updated_at`        datetime(6)        not null,

    primary key (`id`),
    unique key `idx_tempSpaID_name` (`template_space_id`, `name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `template_releases`
(
    `id`                bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `template_name`     varchar(255)       not null,
    `release_name`      varchar(255)       not null,
    `release_memo`      varchar(256) default '',
    `path`              varchar(255)       not null,
    `user`              varchar(64)        not null,
    `user_group`        varchar(64)        not null,
    `privilege`         varchar(64)        not null,
    `signature`         varchar(64)        not null,
    `byte_size`         bigint(1) unsigned not null,

    # Attachment is attachment info of the resource
    `biz_id`            bigint(1) unsigned not null,
    `template_space_id` bigint(1) unsigned not null,
    `template_id`       bigint(1) unsigned not null,

    # CreatedRevision is reversion info of the resource being created.
    `creator`           varchar(64)        not null,
    `created_at`        datetime(6)        not null,

    primary key (`id`),
    unique key `idx_tempSpaID_tempName_path` (`template_space_id`, `template_name`, `path`),
    unique key `idx_tempID_relName` (`template_id`, `release_name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `template_sets`
(
    `id`                   bigint(1) unsigned not null,

    # Spec is specifics of the resource defined with user
    `name`                 varchar(255)       not null,
    `memo`                 varchar(256) default '',
    `template_release_ids` json               not null,

    # Attachment is attachment info of the resource
    `biz_id`               bigint(1) unsigned not null,
    `template_space_id`    bigint(1) unsigned not null,

    # Revision record revision info of the resource
    `creator`              varchar(64)        not null,
    `reviser`              varchar(64)        not null,
    `created_at`           datetime(6)        not null,
    `updated_at`           datetime(6)        not null,

    primary key (`id`),
    unique key `idx_tempSpaID_name` (`template_space_id`, `name`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generators(resource, max_id, updated_at)
values ('template_spaces', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('templates', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('template_releases', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('template_sets', 0, now());