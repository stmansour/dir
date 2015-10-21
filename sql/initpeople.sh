#!/bin/bash

ABSOLUTE_PATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd)/`basename "${BASH_SOURCE[0]}"`
DIR=$(dirname "${ABSOLUTE_PATH}")
echo "LOAD DATA INFILE '${DIR}/peopledata.csv' into table people fields terminated by ',' enclosed by '\"' lines terminated by '\n';" > initpeople.sql
echo "GRANT ALL PRIVILEGES ON Accord TO '${USER}'@'localhost';" >>initpeople.sql
