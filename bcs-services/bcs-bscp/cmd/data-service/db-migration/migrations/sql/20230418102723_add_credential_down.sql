drop table if exists `credentials`;
drop table if exists `credential_scopes`;

delete from id_generators where resource = 'credentials';
delete from id_generators where resource = 'credential_scopes';