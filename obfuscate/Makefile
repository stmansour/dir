all: clean obfuscate

clean:
	rm -f obfuscate

obfuscate:
	go vet
	golint
	go build

