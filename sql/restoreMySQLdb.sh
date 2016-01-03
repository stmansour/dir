#!/bin/bash
# Param 1 = database name 
# Param 2 = mysqldump backup file name
DB=$1
DBfile=$2

echo "DROP DATABASE IF EXISTS ${DB}; CREATE DATABASE ${DB}; USE ${DB};" > restore.sql
echo "source ${DBfile}" >> restore.sql
echo "GRANT ALL PRIVILEGES ON accord TO 'ec2-user'@'localhost' WITH GRANT OPTION;" >> restore.sql
mysql < restore.sql
