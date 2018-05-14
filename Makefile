default: bin

.PHONY: std-bin
std-bin: clean
	go build

.PHONY: bin
bin: std-bin
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
