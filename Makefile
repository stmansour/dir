all: clean dir
	
clean:
	cd obfuscate;make clean
	cd sql;make clean
	rm -f dir

dir:
	go vet
	golint
	go build

newdb:
	cd sql;make;cd ..;./dir

testdb:
	cd sql;make restoredb
