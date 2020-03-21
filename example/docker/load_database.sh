#!/usr/bin/env bash

docker exec -it go-mysql-kafka-db bash -c 'cd /tmp/test_db && mysql -uroot -padmin < employees.sql && mysql -uroot -padmin -t < test_employees_sha.sql'