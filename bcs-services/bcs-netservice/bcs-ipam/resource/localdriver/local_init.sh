#!/bin/bash

sqlite3 bcs-ipam.db '
create table if not exists Resource (
Host varchar(20) not null,
Net varchar(20) not null,
Mask int,
Gateway varchar(20),
Status varchar(10) default "available",
Container varchar(64) default "",
primary key(Host)		
);
'