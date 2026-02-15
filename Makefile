BINARY = sid-synth
GOFLAGS = -v

.PHONY: build render run vet clean install

build:
	go build $(GOFLAGS) -o $(BINARY) .

render:
	go build $(GOFLAGS) -o render ./cmd/render

run: build
	./$(BINARY) $(MIDI)

vet:
	go vet ./...

clean:
	rm -f $(BINARY) render

install:
	go install $(GOFLAGS) .
