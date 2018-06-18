CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-rasterzen"
	@GOPATH=$(GOPATH) go get -u "github.com/akrylysov/algnhsa"
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-cache src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-cache-s3 src/github.com/whosonfirst/

# if you're wondering about the 'rm -rf' stuff below it's because Go is
# weird... https://vanduuren.xyz/2017/golang-vendoring-interface-confusion/
# (20170912/thisisaaronland)

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/rasterd-lambda cmd/rasterd-lambda.go
