all: clean
	./initpeople.sh
	mysql <dir.sql
	#mysql <dir.sql >dbcreate.log 2>&1

clean:
	rm -f initpeople.sql dbcreate.log

backupdb:
	mysqldump accord > testdb.sql

restoredb:
	./restoreMySQLdb.sh accord testdb.sql

install:
	cp restoreMySQLdb.sh /usr/local/accord/testtools/
	@echo "NOTICE:  STILL NEED TO RUN the Accord packager and publisher"

