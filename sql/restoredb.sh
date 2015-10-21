#!/bin/bash
DB="test_db.sql"
if [ "x$1" != "x" ]; then
	DB=$1
fi

echo "DROP DATABASE IF EXISTS accord; CREATE DATABASE accord; USE accord;" > restore.sql
echo "source ${DB}" >> restore.sql
echo "GRANT ALL PRIVILEGES ON Accord TO '${USER}'@'localhost';" >> restore.sql
mysql < restore.sql
