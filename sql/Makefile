all: clean
	./initpeople.sh
	mysql <dir.sql >dbcreate.log 2>&1

clean:
	rm -f initpeople.sql peopledata.csv dbcreate.log

backupdb:
	mysqldump accord > accord_db.sql

restoredb:
	./restoredb.sh test_db.sql
