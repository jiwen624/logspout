default: bin

.PHONY: bin
bin:
	go build *.go

.PHONY: clean
clean:
	rm -f logspout

.PHONY: test
test:
	go test
