


all:
	go build
	go build cmd/bzFlow/bzFlow.go
	go build cmd/bzWhyIgnored/bzWhyIgnored.go

clean:
	rm -f bzFlow bzWhyIgnored

run:
	time go run cmd/bzFlow/bzFlow.go

.PHONY: test
test:
	go test -v

bench:
	go test -v -bench .