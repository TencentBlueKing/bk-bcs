drop table if exists `credentials`;
drop table if exists `credential_scopes`;

delete from id_generators where id = 20;
delete from id_generators where id = 21;