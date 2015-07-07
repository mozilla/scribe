PROJS = scribe scribecmd
GO = GOPATH=$(shell pwd):$(shell go env GOROOT)/bin go

all: $(PROJS)

scribe:
	$(GO) build scribe
	$(GO) install scribe

scribecmd:
	$(GO) install scribecmd

clean:
	rm -rf pkg
