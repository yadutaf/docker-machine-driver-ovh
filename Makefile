PKG = "docker-machine-driver-ovh"
DEPS = $(shell go list -f '{{range .Imports}}{{.}} {{end}}' ./... | tr ' ' '\n' | grep "github.com" | grep -v $(PKG) | sort | uniq | tr '\n' ' ')

vendor:
	GO15VENDOREXPERIMENT=1
	go get -u github.com/FiloSottile/gvt
	for dep in $(DEPS); do            \
		$$GOPATH/bin/gvt fetch $$dep; \
	done

.PHONY: vendor
