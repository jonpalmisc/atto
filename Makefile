build:
	go build -o atto cmd/atto/main.go

run: build
	./atto

install: build
	cp atto /usr/local/bin

uninstall:
	rm /usr/local/bin/atto

clean:
	rm atto
