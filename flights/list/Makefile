.PHONY: build clean deploy test remove

build: test
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/v1 v1/*.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

remove: 
	sls remove -v

deploy: clean build
	sls deploy -v

test:
	go test -v ./...

