BINARY = sid-synth
GOFLAGS = -v

.PHONY: build run vet clean install

build:
	go build $(GOFLAGS) -o $(BINARY) .

run: build
	./$(BINARY) $(MIDI)

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

install:
	go install $(GOFLAGS) .
