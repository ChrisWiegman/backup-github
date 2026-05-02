PKG          := github.com/ChrisWiegman/backup-github
VERSION      := $(shell git describe --tags || echo "0.0.1")
TIMESTAMP    := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
ARGS          = `arg="$(filter-out $@,$(MAKECMDGOALS))" && echo $${arg:-${1}}`

.PHONY: changelog
changelog:
	@if [ ! -d node_modules ]; then npm ci; fi
	npx changie batch $(call ARGS,defaultstring)
	npx changie merge

.PHONY: change
change:
	@if [ ! -d node_modules ]; then npm ci; fi
	npx changie new

.PHONY: clean
clean:
	rm -rf \
		dist \
		vendor \
		node_modules

.PHONY: install
install:
	go mod vendor
	go install \
		-ldflags "-s -w -X $(PKG)/internal/backup.Version=$(VERSION) -X $(PKG)/internal/backup.Timestamp=$(TIMESTAMP)" \
		./cmd/...

.PHONY: lint
lint:
	@if [ ! -f $GOPATH/bin/gilangci-lint  ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;\
	fi
	@golangci-lint \
			run

.PHONY: test
test:
	go \
		test \
		-v \
		-timeout 30s\
		-cover \
		./...

.PHONY: update
update:
	go get -u ./...