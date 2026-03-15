RECIPES_MANAGER_APP	= recipes-manager
TMPAPP  			= $(RECIPES_MANAGER_APP)-tmp
SNAPSHOT			= $(RECIPES_MANAGER_APP)-snapshot
GO_VERSION     		= $(shell go version)
DATE     			= $(shell date +%F_%T)

RECIPES_MANAGER_VERSION		= $(shell git describe --tags --always --match=v* 2> /dev/null || echo v0.0.0)
RECIPES_MANAGER_GIT_HASH	= $(shell git rev-parse --short HEAD)

RECIPES_MANAGER_REPO 				?= github.com/ottenwbe/recipes-manager
RECIPES_MANAGER_DOCKER_SHOULD_PUSH	?= false
RECIPES_MANAGER_DOCKER_PREFIX  		?= ottenwbe
RECIPES_MANAGER_MAINTAINER			?= Beate Ottenwaelder <ottenwbe.public@gmail.com>

VERSIONPKG = "$(RECIPES_MANAGER_REPO)/core.appVersionString"

DOCKER_REGISTRY ?= docker.io

GO      = go
GOFMT   = gofmt
GOVET   = go vet

M = $(shell printf "\033[34;1m▶\033[0m")

DOCKER_NETWORK_NAME				?= recipes-manager-net
DB_CONTAINER_NAME				?= db-recipes-manager
APP_CONTAINER_NAME				?= backend-recipes-manager

RECIPES_MANAGER_BUILD_DOCKER_HOST	?=

RECIPES_MANAGER_DOCKER_IMAGE	= $(DOCKER_REGISTRY)/$(RECIPES_MANAGER_DOCKER_PREFIX)/$(RECIPES_MANAGER_APP)
RECIPES_MANAGER_DOCKER_PARAMS	= \
								--build-arg "APP=$(RECIPES_MANAGER_APP)-$(RECIPES_MANAGER_VERSION)" \
								--label "version=$(RECIPES_MANAGER_VERSION)" \
								--label "go=$(GO_VERSION)" \
								--label "build_date=$(DATE)" \
								--label "maintaner=$(RECIPES_MANAGER_MAINTAINER)" \
								--label "git-hash=$(RECIPES_MANAGER_GIT_HASH)" \
								--label "git-repo=$(RECIPES_MANAGER_REPO)"

.PHONY: release
release: ; $(info $(M) building executable…) @ ## Build the app's binary release version
	@$(GO) build \
		-tags release \
		-ldflags "-s -w -X $(VERSIONPKG)=$(RECIPES_MANAGER_VERSION)" \
		-o $(RECIPES_MANAGER_APP)-$(RECIPES_MANAGER_VERSION) \
		*.go

.PHONY: snapshot
snapshot:  ; $(info $(M) building snapshot…) @ ## Build the app's snapshot version
		@$(GO) build \
		-o $(SNAPSHOT) \
		-ldflags "-X $(VERSIONPKG)=$(RECIPES_MANAGER_VERSION)" \
		*.go

.PHONY: start
start: fmt ; $(info $(M) running the app locally…) @ ## Run the program's snapshot version
	@$(GO) build \
	    -o $(TMPAPP) \
    	-ldflags "-X $(VERSIONPKG)=$(RECIPES_MANAGER_VERSION)" \
    	*.go && ./$(TMPAPP)

# Quality and Testing

.PHONY: verify
verify: mod-verify vet test; $(info $(M) QA steps…) @ ## Run all QA steps
	@echo "End of QA steps..."

.PHONY: mod-verify
mod-verify: ; $(info $(M) verifying modules…) @ ## Run go mod verify
	@$(GO) mod verify

.PHONY: vet
vet: ; $(info $(M) running vet…) @ ## Run go vet
	@$(GOVET) ./...

.PHONY: sbom
sbom: ; $(info $(M) creating SBOM...) @ ## Create SBOM
	@$(GO) run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest app -json -output recipes-manager.bom.json .

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@$(GOFMT) -s -l -w .

.PHONY: test
test: ; $(info $(M) running tests…) @ ## Run tests
	@sh test.sh

# Misc

.PHONY: docker-arm
docker-arm: ; $(info $(M) building arm docker image...) @  ## Create docker image for arm
ifndef RECIPES_MANAGER_BUILD_DOCKER_HOST
	docker build $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION) -f Dockerfile.armhf .
else
	docker -H $(RECIPES_MANAGER_BUILD_DOCKER_HOST)  build $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION) -f Dockerfile.armhf .
endif

.PHONY: docker
docker: ; $(info $(M) building docker image...) @ ## Create docker image
ifndef RECIPES_MANAGER_BUILD_DOCKER_HOST
	docker build $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION) -f Dockerfile .
else
	docker -H $(RECIPES_MANAGER_BUILD_DOCKER_HOST) $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION) -f Dockerfile .
endif

.PHONY: docker-dev
docker-dev: ; $(info $(M) building development docker image...) @ ## Create docker image for development
ifndef RECIPES_MANAGER_BUILD_DOCKER_HOST	
	docker build $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):development -f Dockerfile .
else
	docker -H $(RECIPES_MANAGER_BUILD_DOCKER_HOST) build  $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):development -f Dockerfile .
endif

.PHONY: docker-network
docker-network:
	@docker network inspect $(DOCKER_NETWORK_NAME) >/dev/null 2>&1 || docker network create $(DOCKER_NETWORK_NAME)

.PHONY: docker-start-db
docker-start-db: docker-network ; $(info $(M) starting mongodb...) @ ## Start mongodb container
	@docker run -d --name=$(DB_CONTAINER_NAME) --network=$(DOCKER_NETWORK_NAME) -p 27018:27017 mongo:8
	@echo "Waiting for MongoDB to be ready..."
	@until docker exec $(DB_CONTAINER_NAME) mongosh --port 27017 --eval "db.adminCommand('ping')" >/dev/null 2>&1; do sleep 1; done

.PHONY: docker-start
docker-start: docker-dev docker-start-db ; $(info $(M) starting docker containers...) @ ## Build and start dev containers (app and db)
	@docker run -d --name=$(APP_CONTAINER_NAME) --network=$(DOCKER_NETWORK_NAME) -p 8080:8080 \
		-e GO_COOK_RECIPEDB_HOST=mongodb://$(DB_CONTAINER_NAME):27017 \
		$(RECIPES_MANAGER_DOCKER_IMAGE):development

.PHONY: docker-stop-db
docker-stop-db: ; $(info $(M) stopping mongodb...) @ ## Stop and remove mongodb container
	@docker stop $(DB_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(DB_CONTAINER_NAME) >/dev/null 2>&1 || true

.PHONY: docker-stop
docker-stop: ; $(info $(M) stopping docker containers...) @ ## Stop and remove running dev containers
	@docker stop $(APP_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(APP_CONTAINER_NAME) >/dev/null 2>&1 || true
	@$(MAKE) docker-stop-db

.PHONY: docker-login
docker-login: ; $(info $(M) login to docker hub...) @ ## Login to Dockerhub
	echo $(DOCKER_PASSWORD) | docker login -u $(DOCKER_USERNAME) $(DOCKER_REGISTRY) --password-stdin

.PHONY: docker-push-dev
docker-push-dev: ; $(info $(M) push snapshot to registry...) @ ## Push docker image with a development version
	docker push $(RECIPES_MANAGER_DOCKER_IMAGE):development --tls-verify=false 

.PHONY: dockerx
dockerx: ; ## Build docker image with buildx
	docker buildx build --output "type=image,push=$(RECIPES_MANAGER_DOCKER_SHOULD_PUSH)" --platform linux/arm64/v8,linux/amd64 $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION) -f Dockerfile .

.PHONY: dockerx-dev
dockerx-dev:  ; ## Build development docker image with buildx
	docker buildx build --output "type=image,push=$(RECIPES_MANAGER_DOCKER_SHOULD_PUSH)" --platform linux/arm64/v8,linux/amd64 $(RECIPES_MANAGER_DOCKER_PARAMS) -t $(RECIPES_MANAGER_DOCKER_IMAGE):development -f Dockerfile .

.PHONY: docker-push
docker-push: ; ## Push docker image
ifndef RECIPES_MANAGER_BUILD_DOCKER_HOST
	docker push $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION)
else
	docker -H $(RECIPES_MANAGER_BUILD_DOCKER_HOST) push $(RECIPES_MANAGER_DOCKER_IMAGE):$(RECIPES_MANAGER_VERSION)
endif

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy
ifneq (,$(wildcard vendor))
	go mod vendor
endif

.PHONY: api-docu
api-docu: ## Create the API documentation
	@swag init --exclude vendor

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(RECIPES_MANAGER_VERSION)

.PHONY: details
details:
	@echo "APP:    " $(RECIPES_MANAGER_APP) $(RECIPES_MANAGER_VERSION) $(VERSIONPKG)
	@echo "GO:     " $(GO_VERSION)
	@echo "GIT:    " $(RECIPES_MANAGER_REPO) $(RECIPES_MANAGER_GIT_HASH)
	@echo "DOCKER: " $(RECIPES_MANAGER_DOCKER_PARAMS)

.PHONY: date
date:
	@echo $(DATE)

.PHONY: ls
ls:
	@echo $(GOPATH)
