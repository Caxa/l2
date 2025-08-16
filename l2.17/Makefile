BINARY = telnet-client
ARGS ?= google.com 80   

.PHONY: all build build-race run run-race clean

all: build


build:
	go build -o $(BINARY) main.go


build-race:
	go build -race -o $(BINARY) main.go


run: build
	./$(BINARY) $(ARGS)


run-race: build-race
	./$(BINARY) $(ARGS)


clean:
	rm -f $(BINARY)
