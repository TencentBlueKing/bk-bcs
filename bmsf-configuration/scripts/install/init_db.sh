#!/usr/bin/env bash

set -eo pipefail

source ./bscp.env

# database
MYSQL="mysql -h${DB_HOST} -u${DB_USER} -P${DB_PORT} -p${DB_PASSWD} --default-character-set=utf8mb4 -A -N"

# tables
SQL="bscp.sql"

# init database
$MYSQL -e "source ${SQL}"
