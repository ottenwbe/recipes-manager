APP      = go-cook
TMPAPP   = $(APP)-tmp
SNAPSHOT = $(APP)-snapshot
DATE     = $(shell date +%F_%T)
VERSION  = $(shell git describe --tags --always --match=v* 2> /dev/null || echo v0.0.0)

MAINTAINER ?= Beate Ottenwaelder <ottenwbe.public@gmail.com>

VERSIONPKG = "github.com/ottenwbe/go-life/core.appVersionString"

GO      = go
GOFMT   = gofmt
GOVET   = go vet
GOLINT  = golint

M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: release
release: ; $(info $(M) building executable…) @ ## Build the app's binary release version
	@$(GO) build \
		-tags release \
		-ldflags "-s -w" \
		-ldflags "-X $(VERSIONPKG)=$(VERSION)" \
		-o $(APP)-$(VERSION) \
		*.go

.PHONY: build
build:  ; $(info $(M) building snapshot…) @ ## Build the app's snapshot version
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

.PHONY: verify
verify: fmt mod-verify vet lint test; $(info $(M) QA steps…) @ ## Run all QA steps
	@echo "End of QA steps..."

.PHONY: mod-verify
mod-verify: ; $(info $(M) verifying modules…) @ ## Run go mod verify
	@$(GO) mod verify

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

.PHONY: test
test: ; $(info $(M) running tests…) @ ## Run tests
	@sh test.sh

# Misc

.PHONY: docker-arm
docker-arm: ## Create docker image for arm
ifndef BUILD_DOCKER_HOST
	docker build --label "version=${VERSION}" --build-arg "APP=$(APP)-$(VERSION)"  --label "build_date=${DATE}" --label "maintaner=$(MAINTAINER)" -t $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION) -f Dockerfile.armhf .
else
	docker -H $(BUILD_DOCKER_HOST)  build --build-arg "APP=$(APP)-$(VERSION)"  --label "version=${VERSION}" --label "build_date=${DATE}" --label "maintaner=$(MAINTAINER)" -t $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION) -f Dockerfile.armhf .
endif

.PHONY: docker
docker: ## Create docker image
ifndef BUILD_DOCKER_HOST
	docker build --label "version=$(VERSION)" --build-arg "APP=$(APP)-$(VERSION)"  --label "build_date=$(DATE)"  --label "maintaner=$(MAINTAINER)" -t $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION) -f Dockerfile .
else
	docker -H $(BUILD_DOCKER_HOST) build --build-arg "APP=$(APP)-$(VERSION)"  --label "version=$(VERSION)" --label "build_date=$(DATE)"  --label "maintaner=$(MAINTAINER)" -t $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION) -f Dockerfile .
endif

.PHONY: docker-push
docker-push: ## Push docker image
ifndef BUILD_DOCKER_HOST
	echo $(DOCKER_PASSWORD) | docker login -u $(DOCKER_USERNAME) --password-stdin; \
	docker push $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION)
else
	docker -H $(BUILD_DOCKER_HOST) push $(DOCKER_PREFIX)go-cook:$(GO_COOK_ARCH)-$(VERSION)
endif

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