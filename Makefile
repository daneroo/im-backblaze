


all:
	go build
	go build cmd/bzFlow/bzFlow.go
	go build cmd/bzWhyIgnored/bzWhyIgnored.go

clean:
	rm -f bzFlow bzWhyIgnored

sample:
	grep -h '"/"' raw-tx-2018-*.jsonl >sample.jsonl
	grep -h '"/Volumes/Space/"' raw-tx-2018-*.jsonl >>sample.jsonl
	grep -h '"/Users/daniel/"' raw-tx-2018-*.jsonl >>sample.jsonl
	grep -h '"/Users/daniel/Library/Containers/com.docker.docker/Data/"' raw-tx-2018-*.jsonl >>sample.jsonl
	

run:
	time go run cmd/bzFlow/bzFlow.go

.PHONY: test
test:
	go test -v

bench:
	go test -v -bench .