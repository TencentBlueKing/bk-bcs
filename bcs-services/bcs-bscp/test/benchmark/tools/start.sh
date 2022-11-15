#!/bin/bash

# bench test env address.
cache_host=$(eval echo '$'ENV_BENCH_TEST_CACHE_REQUEST_HOST)
feed_host=$(eval echo '$'ENV_BENCH_TEST_FEED_REQUEST_HOST)

# Warn: mysql used to clear the mysql of the bench test environment.
mysql_ip=$(eval echo '$'ENV_BENCH_TEST_MYSQL_IP)
mysql_port=$(eval echo '$'ENV_BENCH_TEST_MYSQL_PORT)
mysql_user=$(eval echo '$'ENV_BENCH_TEST_MYSQL_USER)
mysql_password=$(eval echo '$'ENV_BENCH_TEST_MYSQL_PW)
mysql_db=$(eval echo '$'ENV_BENCH_TEST_MYSQL_DB)

# bench result statistics result save file path.
output_path=$(eval echo '$'ENV_BENCH_TEST_OUTPUT_PATH)

./cache_service.test -host ${cache_host} -test.run TestReport -concurrent 100 -sustain-seconds 10 -output-path ./cache_report.html
./feed_server_http.test -host ${feed_host} -test.run TestReport -mysql-ip=${mysql_ip} -mysql-port=${mysql_port} -mysql-user=${mysql_user} -mysql-passwd=${mysql_password} -mysql-db=${mysql_db} -concurrent 100 -sustain-seconds 60 -output-path ./feed_report.html

cat ./cache_report.html ./feed_report.html > ${output_path}
