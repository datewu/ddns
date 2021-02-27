#GO11MODULES=on
APP?=$(shell basename `pwd`)
REGISTRY?=gcr.io/images
VER=$(shell git describe --abbrev=0 --tag || echo v0.0.0) # `echo xxx`
COMMIT_SHA=$(shell git rev-parse --short HEAD)
LD_FLAGS="-s -w -X main.GitCommit=${COMMIT_SHA} -X main.SemVer=${VER}"

.PHONY: build
## build: build the application
build: clean 
	@echo "Building..."
	@go build  \
	-race \
	-gcflags=all=-d=checkptr \
	-ldflags ${LD_FLAGS} \
	-o ${APP} 
	
dlv-debug: clean
	@echo "Building for delve debug..."
	@go build \
	-ldflags ${LD_FLAGS} \
	-ldflags=-compressdwarf=false \
	-gcflags=all=-d=checkptr \
	-gcflags="all=-N -l" \
	-o ${APP} 

.PHONY: dev
dev: build
	. .env && \
	./${APP} 


.PHONY: run
## run: runs go run main.go
run:
	go run -race main.go

.PHONY: clean
## clean: cleans the binary
clean:
	@echo "Cleaning"
	@rm -rf ${APP}

.PHONY: test
## test: runs go test with default values
test:
	JWT_SECRET_KEY="qq@oshlkol" \
	go test -timeout 300s -v -count=1 -race ./...

.PHONY: update
## update: runs go get -u 
update:
	go get -u ./...
	go mod tidy

.PHONY: custom
## custom: populate the template
custom:
	@rm -rf go.*
	@echo ${APP} > .gitignore
	@echo .env >> .gitignore
	@echo .DS_Store >> .gitignore
	@echo Hello ${APP} `date` > README.md
	@go mod init github.com/datewu/${APP}
	@go build
	@go test
	@-go get -u
	#@make build
	#@make test
## add github secrets settings
	@git add .
	@git commit -am "init custom"
	@-git push 
	echo ================
	echo before do any git tag v0.0.x, please
	echo change CHANGE-ME in .github/workflows/docker.yml 
	echo ================

.PHONY: build-tokenizer
## build-tokenizer: build the tokenizer application
build-tokenizer:
	${MAKE} -c tokenizer build

.PHONY: setup
## setup: setup go modules
setup:
	@go mod init \
		&& go mod tidy \
		&& go mod vendor
	
# helper rule for deployment
check-environment:
ifndef ENV
	$(error ENV not set, allowed values - `staging` or `production`)
endif

.PHONY: docker-build
## docker-build: builds the stringifier docker image to registry
docker-build: build
	docker build -t ${APP}:${COMMIT_SHA} .

.PHONY: docker-push
## docker-push: pushes the stringifier docker image to registry
docker-push: check-environment docker-build
	docker push ${REGISTRY}/${ENV}/${APP}:${COMMIT_SHA}

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
