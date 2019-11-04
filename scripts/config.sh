#!/bin/bash


export bcs_home="/data/bcs"
export BCS_HOME="${bcs_home}"

# bcs common
export log_dir="${bcs_home}/logs/bcs"
export pid_dir="/var/run/bcs"

export caFile="${bcs_home}/cert/bcs/test-ca.crt"
export serverCertFile="${bcs_home}/cert/bcs/test-server.crt"
export serverKeyFile="${bcs_home}/cert/bcs/test-server.key"
export clientCertFile="${bcs_home}/cert/bcs/test-client.crt"
export clientKeyFile="${bcs_home}/cert/bcs/test-client.key"

export service_etcd_cert="${bcs_home}/cert/etcd/etcd.pem"
export service_etcd_key="${bcs_home}/cert/etcd/etcd-key.pem"
export service_etcd_ca="${bcs_home}/cert/etcd/ca.pem"

# bcs-api specified configuration
export bcsApiPort=30443
export bcsApiMesosWebconsolePort=8083
export bcsApiMetricPort=8082
export apiEdition=opensource
## auth
export bkiamAuthHost="http://auth.test.com"
## bke
export coreDatabaseDsn="testmysqlDsn"
export adminUser="user"
export adminToken="testToken"

# bcs-storage
export bcsStoragePort=50024
export bcsStorageMetricPort=50025
export storageDbConfig="./storage-database.conf"
export mongodbHost="test1:7018,test2:7018,test3:7018"
export mongodbUsername="storage"
export mongodbPassword="storagePassword"
export mongodbOplogCollection="oplog.rs"
export ConfigDbHost="db1,db2,db3"
export ConfigDbUsername="storage"
export ConfigDbPassword="storagepassword"