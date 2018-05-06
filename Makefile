default: bin

.PHONY: bin
bin: clean-bin
	go build

.PHONY: tiny
tiny: bin
	upx logspout

.PHONY: clean-bin
clean-bin:
	rm -f logspout

.PHONY: clean-logs
clean-logs:
	rm -f *.log

.PHONY: clean
clean: clean-bin clean-logs

.PHONY: test
test:
	go test
