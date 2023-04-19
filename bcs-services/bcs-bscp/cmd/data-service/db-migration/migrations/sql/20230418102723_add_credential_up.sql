insert into id_generators(resource, max_id, updated_at)
values ('credentials', 0, now());
insert into id_generators(resource, max_id, updated_at)
values ('credential_scopes', 0, now());

CREATE TABLE if not exists `credentials` (
    `id` bigint(1) unsigned NOT NULL,
    `biz_id` bigint(1) unsigned NOT NULL,
    `credential_type` varchar(64) NOT NULL,
    `enc_credential` varchar(64) NOT NULL,
    `enc_algorithm` varchar(64) NOT NULL,
    `memo` longtext,
    `enable` tinyint unsigned   default 0,
    `creator` varchar(64) NOT NULL,
    `reviser` varchar(64) NOT NULL,
    `created_at` datetime(6) NOT NULL,
    `updated_at` datetime(6) NOT NULL,
    `expired_at` datetime(6) NOT NULL,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE if not exists `credential_scopes` (
     `id` bigint(1) unsigned NOT NULL,
     `biz_id` bigint(1) unsigned NOT NULL,
     `credential_id` bigint(1) unsigned NOT NULL,
     `credential_scope` varchar(64) NOT NULL,
     `creator` varchar(64) NOT NULL,
     `reviser` varchar(64) NOT NULL,
     `updated_at` datetime(6) NOT NULL,
     `created_at` datetime(6) NOT NULL,
     `expired_at` datetime(6) NOT NULL,
     PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;