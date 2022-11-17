#!/bin/bash

# suite test env address.
api_host=$(eval echo '$'ENV_SUITE_TEST_API_REQUEST_HOST)
cache_host=$(eval echo '$'ENV_SUITE_TEST_CACHE_REQUEST_HOST)
feed_host=$(eval echo '$'ENV_SUITE_TEST_FEED_REQUEST_HOST)

# Warn: mysql used to clear the mysql of the suite test environment.
mysql_ip=$(eval echo '$'ENV_SUITE_TEST_MYSQL_IP)
mysql_port=$(eval echo '$'ENV_SUITE_TEST_MYSQL_PORT)
mysql_user=$(eval echo '$'ENV_SUITE_TEST_MYSQL_USER)
mysql_password=$(eval echo '$'ENV_SUITE_TEST_MYSQL_PW)
mysql_db=$(eval echo '$'ENV_SUITE_TEST_MYSQL_DB)

# sidecar_start_cmd sidecar start cmd.
sidecar_start_cmd=$(eval echo '$'ENV_SUITE_TEST_SIDECAR_START_CMD)

# go test result export json file save dir.
save_dir=$(eval echo '$'ENV_SUITE_TEST_SAVE_DIR)
# go test result statistics result save file path.
output_path=$(eval echo '$'ENV_SUITE_TEST_OUTPUT_PATH)

# exec go test.
if [ -d "${save_dir}" ]; then
  rm -r ${save_dir}
fi

mkdir ${save_dir}
./api.test -test.run TestApi -api-host=${api_host} -mysql-ip=${mysql_ip} -mysql-port=${mysql_port} -mysql-user=${mysql_user} -mysql-passwd=${mysql_password} -mysql-db=${mysql_db} -convey-json=true > ${save_dir}/api.json
./cache.test -cache-host=${cache_host} -api-host=${api_host} -mysql-ip=${mysql_ip} -mysql-port=${mysql_port} -mysql-user=${mysql_user} -mysql-passwd=${mysql_password} -mysql-db=${mysql_db} -convey-json=true > ${save_dir}/cache.json
./feed.test  -api-host=${api_host} -feed-host=${feed_host} -mysql-ip=${mysql_ip} -mysql-port=${mysql_port} -mysql-user=${mysql_user} -mysql-passwd=${mysql_password} -mysql-db=${mysql_db} -convey-json=true > ${save_dir}/feed.json
./sidecar.test  -api-host=${api_host} -feed-host=${feed_host} -mysql-ip=${mysql_ip} -mysql-port=${mysql_port} -mysql-user=${mysql_user} -mysql-passwd=${mysql_password} -mysql-db=${mysql_db} -sidecar-start-cmd "${sidecar_start_cmd}" -convey-json=true > ${save_dir}/sidecar.json

# statistics.
./tools.sh -input-dir=${save_dir} -output-path=${output_path}
