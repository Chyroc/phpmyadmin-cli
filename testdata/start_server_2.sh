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

docker run --name mysql_db_server_1 -e MYSQL_ROOT_PASSWORD=pass -d -p 3307:3306 mysql
docker run --name mysql_db_server_2 -e MYSQL_ROOT_PASSWORD=pass -d -p 3308:3306 mysql

retry_curl 3307
retry_curl 3308

docker run --name myadmin -d --link mysql_db_server_1:db_1 --link mysql_db_server_2:db_2 -e PMA_PORTS=3306,3306 -e PMA_HOSTS=db_1,db_2 -p 8000:80 phpmyadmin/phpmyadmin

retry_curl 8000
