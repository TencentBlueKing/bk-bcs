drop table if exists `template_spaces`;
drop table if exists `templates`;
drop table if exists `template_releases`;
drop table if exists `template_sets`;

delete from id_generators where resource = 'template_spaces';
delete from id_generators where resource = 'templates';
delete from id_generators where resource = 'template_releases';
delete from id_generators where resource = 'template_sets';