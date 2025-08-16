BINARY = minishell

.PHONY: all build build-race run run-race clean

all: build

build:
	go build -o $(BINARY) main.go

build-race:
	go build -race -o $(BINARY) main.go

run: build
	./$(BINARY)

run-race: build-race
	./$(BINARY)

clean:
	rm -f $(BINARY)
