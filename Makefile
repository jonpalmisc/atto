build:
	go build -o atto cmd/atto/main.go

run:
	go build -o atto cmd/atto/main.go && ./atto

clean:
	rm atto
