APP		 = go-cook
TMPAPP	 = $(APP)-tmp
SNAPSHOT = $(APP)-snapshot
DATE     = $(shell date +%F_%T)
VERSION  = $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0.0.0)

VERSIONPKG = "github.com/ottenwbe/go-life/core.appVersionString"

GO      = go
GOFMT   = gofmt
GOVET   = go vet
GOLINT  = golint

M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: release
release: fmt vet lint ; $(info $(M) building executable…) @ ## Build the app's binary release version
	@$(GO) build \
		-tags release \
		-ldflags "-s -w" \
		-ldflags "-X $(VERSIONPKG)=$(VERSION)" \
		-o $(APP) \
		*.go

.PHONY: build
build: fmt vet lint ; $(info $(M) building snapshot…) @ ## Build the app's snapshot version
	@$(GO) build \
		-o $(SNAPSHOT) \
		-ldflags "-X $(VERSIONPKG)=$(VERSION)" \
		*.go

.PHONY: start
start: fmt ; $(info $(M) running the app locally…) @ ## Run program's snapshot version
	@$(GO) build \
	    -o $(TMPAPP) \
    	-ldflags "-X $(VERSIONPKG)=$(VERSION)" \
    	*.go && ./$(TMPAPP)

# Quality and Testing

.PHONY: vet
vet: ; $(info $(M) running vet…) @ ## Run go vet
	@for d in $$($(GO) list ./...); do \
		$(GOVET) $${d};  \
	done

.PHONY: lint
lint: ; $(info $(M) running golint…) @ ## Run golint
	@for d in $$($(GO) list ./...); do \
		$(GOLINT) $${d};  \
	done

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
		$(GOFMT) -l -w $$d/*.go  ; \
	 done

cleanup: ; docker stop test-mongo && docker rm -v test-mongo

.PHONY: test
test: ; $(info $(M) running tests…) @ ## Run tests
	@sh test.sh

# Misc

.PHONY: docker-arm
docker-arm: ## Create docker image
	docker -H $(BUILD_DOCKER_HOST)  build --label "version=m${VERSION}" --label "build_date=${DATE}"  --label "maintaner=Beate Ottenwaelder <ottenwbe.public@gmail.com>" -t $(DOCKER_PREFIX)go-cook:$(VERSION) -f Dockerfile.armhf .

.PHONY: docker-push
	docker-push: ## Push docker image
		docker -H $(BUILD_DOCKER_HOST) push $(DOCKER_PREFIX)go-cook:$(VERSION)

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: date
date:
	@echo $(DATE)

.PHONY: ls
ls:
	@echo $(GOPATH)