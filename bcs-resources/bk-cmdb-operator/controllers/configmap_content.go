/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package controllers

const (
	//apiserver config
	apiserverConfContent = `[auth]
address = http://iam.blueking.com/
appCode = bk_cmdb
appSecret = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

	// audit config
	auditConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true`

	// core config
	coreConfContenetTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// datacollection config
	dataCnConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[snap-redis]
host = {{.RedisHost}}:{{.RedisPort}}
usr = root
pwd = {{.RedisPwd}}
database = 0

[discover-redis]
host = {{.RedisHost}}:{{.RedisPort}}
usr = root
pwd = {{.RedisPwd}}
database = 0

[netcollect-redis]
host = {{.RedisHost}}:{{.RedisPort}}
usr = root
pwd = {{.RedisPwd}}
database = 0

[redis]
host = {{.RedisHost}}:{{.RedisPort}}
usr = root
pwd = {{.RedisPwd}}
database = 0`

	// eventserver config
	eventServerConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// host config
	hostConfContentTemplate = `[gse]
addr={{.ZookeeperHost}}:{{.ZookeeperPort}}
user=bkzk
pwd=L%blKas

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// hostcontroller config
	hostCtrlConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// migrate config
	migrateConfContentTemplate = `[config-server]
addrs={{.ZookeeperHost}}:{{.ZookeeperPort}}
usr=
pwd=

[register-server]
addrs={{.ZookeeperHost}}:{{.ZookeeperPort}}
usr=
pwd=

[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[confs]
dir = /etc/configures/

[errors]
res=conf/errors

[language]
res=conf/language`

	// objectcontroller config
	objectCtrlConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// proc config
	procConfContentTemplate = `[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
port={{.RedisPort}}
database = 0`

	// proccontroller config
	procCtrlConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000`

	// topo config
	topoConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[level]
businessTopoMax=7`

	//txc config
	txcConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000

[transaction]
enable=false
transactionLifetimeSecond=60`

	// webserver config
	webserverConfContentTemplate = `[api]
version=v3

[session]
name=cc3
skip=1
defaultlanguage=zh-cn
host={{.RedisHost}}
port={{.RedisPort}}
secret={{.RedisPwd}}
multiple_owner=0

[site]
domain_url=http://{{.IngressDomain}}
app_code=cc
check_url=http://127.0.0.1:8088/login/accounts/get_user/?bk_token=
bk_account_url=http://127.0.0.1:8088/login/accounts/get_all_user/?bk_token=%s
resources_path=/tmp/
html_root=../web

[app]
agent_app_url=http://127.0.0.1:8088/console/?app=bk_agent_setup`

	// task config
	taskConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[redis]
host={{.RedisHost}}
usr=root
pwd={{.RedisPwd}}
database=0
port={{.RedisPort}}
maxOpenConns=3000
maxIDleConns=1000

[transaction]
enable=false
transactionLifetimeSecond=60`

	// operation config
	operationConfContentTemplate = `[mongodb]
host = {{.MongoHost}}
usr = {{.MongoUsername}}
pwd = {{.MongoPwd}}
database = {{.MongoDatabase}}
port = {{.MongoPort}}
maxOpenConns = 3000
maxIdleConns = 1000
mechanism=SCRAM-SHA-1
enable=true

[timer]
spec = 00:30  # 00:00 - 23:59`
)
