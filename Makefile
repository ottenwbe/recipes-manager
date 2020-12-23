GO_COOK_APP     = go-cook
TMPAPP  		= $(GO_COOK_APP)-tmp
SNAPSHOT		= $(GO_COOK_APP)-snapshot
DATE     		= $(shell date +%F_%T)
GO_COOK_VERSION	= $(shell git describe --tags --always --match=v* 2> /dev/null || echo v0.0.0)
GO_COOK_ARCH	?= default

GO_COOK_MAINTAINER ?= Beate Ottenwaelder <ottenwbe.public@gmail.com>

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
		-ldflags "-X $(VERSIONPKG)=$(GO_COOK_VERSION)" \
		-o $(GO_COOK_APP)-$(GO_COOK_VERSION) \
		*.go

.PHONY: build
build:  ; $(info $(M) building snapshot…) @ ## Build the app's snapshot version
	@$(GO) build \
		-o $(SNAPSHOT) \
		-ldflags "-X $(VERSIONPKG)=$(GO_COOK_VERSION)" \
		*.go

.PHONY: start
start: fmt ; $(info $(M) running the app locally…) @ ## Run program's snapshot version
	@$(GO) build \
	    -o $(TMPAPP) \
    	-ldflags "-X $(VERSIONPKG)=$(GO_COOK_VERSION)" \
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
docker-arm: docker-login ; $(info $(M) building arm docker image...) @  ## Create docker image for arm
ifndef GO_COOK_BUILD_DOCKER_HOST
	docker build --label "version=${GO_COOK_VERSION}" --build-arg "APP=$(GO_COOK_APP)-$(GO_COOK_VERSION)"  --label "build_date=${DATE}" --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH) -f Dockerfile.armhf .
else
	docker -H $(GO_COOK_BUILD_DOCKER_HOST)  build --build-arg "APP=$(GO_COOK_APP)-$(GO_COOK_VERSION)"  --label "version=${GO_COOK_VERSION}" --label "build_date=${DATE}" --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH) -f Dockerfile.armhf .
endif

.PHONY: docker
docker: docker-login ; $(info $(M) building docker image...) @ ## Create docker image
ifndef GO_COOK_BUILD_DOCKER_HOST
	docker build --label "version=$(GO_COOK_VERSION)" --build-arg "APP=$(GO_COOK_APP)-$(GO_COOK_VERSION)"  --label "build_date=$(DATE)"  --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH) -f Dockerfile .
else
	docker -H $(GO_COOK_BUILD_DOCKER_HOST) build --build-arg "APP=$(GO_COOK_APP)-$(GO_COOK_VERSION)"  --label "version=$(GO_COOK_VERSION)" --label "build_date=$(DATE)"  --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH) -f Dockerfile .
endif

.PHONY: docker-snapshot
docker-snapshot: build ; $(info $(M) building docker-snapshot image...) @ ## Create docker image of the snapshot 
ifndef GO_COOK_BUILD_DOCKER_HOST	
	docker build --label "version=$(GO_COOK_VERSION)" --build-arg "APP=$(SNAPSHOT)"  --label "build_date=$(DATE)"  --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:SNAPSHOT-$(GO_COOK_ARCH) -f Dockerfile .
else
	docker -H $(GO_COOK_BUILD_DOCKER_HOST) build --build-arg "APP=$(SNAPSHOT)"  --label "version=$(GO_COOK_VERSION)" --label "build_date=$(DATE)"  --label "maintaner=$(GO_COOK_MAINTAINER)" -t $(GO_COOK_DOCKER_PREFIX)go-cook:SNAPSHOT-$(GO_COOK_ARCH) -f Dockerfile . 
endif

.PHONY: docker-login
docker-login: ; $(info $(M) login to docker hub...) @ ## Login to Dockerhub
	echo $(DOCKER_PASSWORD) | docker login -u $(DOCKER_USERNAME) --password-stdin

.PHONY: docker-push-snapshot
docker-push-snapshot: docker-snapshot docker-login ; $(info $(M) push snapshot to docker hub...) @ ## Push docker image with a development version
	docker push $(GO_COOK_DOCKER_PREFIX)go-cook:SNAPSHOT-$(GO_COOK_ARCH)

.PHONY: docker-push
docker-push: docker-login ; ## Push docker image
ifndef GO_COOK_BUILD_DOCKER_HOST
	docker push $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH)
else
	docker -H $(GO_COOK_BUILD_DOCKER_HOST) push $(GO_COOK_DOCKER_PREFIX)go-cook:$(GO_COOK_VERSION)-$(GO_COOK_ARCH)
endif

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(GO_COOK_VERSION)
.PHONY: date
date:
	@echo $(DATE)

.PHONY: ls
ls:
	@echo $(GOPATH)