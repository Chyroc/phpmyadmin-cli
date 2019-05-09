#!/usr/bin/env bash

set -e

function retry_curl {
while ! curl --connect-timeout 5 --max-time 10 --retry 5 --retry-delay 0 --retry-max-time 10 'http://127.0.0.1:'$1
do
    { echo "Exit status of curl: $?"
      echo "Retrying ..."
    } 1>&2
    sleep 1
done
}

docker run --name mysql_db_server -e MYSQL_ROOT_PASSWORD=pass -d -p 3307:3306 mysql

retry_curl 3307

docker run --name myadmin -d --link mysql_db_server:db -p 8000:80 phpmyadmin/phpmyadmin

retry_curl 8000
