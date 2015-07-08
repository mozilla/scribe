PROJS = scribe scribecmd evrtest
GO = GOPATH=$(shell pwd):$(shell go env GOROOT)/bin go
export SCRIBECMD = $(shell pwd)/bin/scribecmd
export EVRTESTCMD = $(shell pwd)/bin/evrtest

all: $(PROJS)

evrtest:
	$(GO) install evrtest

scribe:
	$(GO) build scribe
	$(GO) install scribe

scribecmd:
	$(GO) install scribecmd

runtests: $(PROJS)
	cd test && $(MAKE) runtests

clean:
	rm -rf pkg
	rm -f bin/*
	cd test && $(MAKE) clean
