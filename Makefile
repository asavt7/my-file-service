DEFAULT_ENV_FILE='.env'
DEV_ENV_FILE='.env.local.dev'
ENV_FILE := $(shell if [ ! -f $(DEFAULT_ENV_FILE) ]; then echo $(DEV_ENV_FILE) ; else echo $(DEFAULT_ENV_FILE) ; fi   )
include $(ENV_FILE)
export

MAIN_APP_FILE=./cmd/main.go
MAIN_APP_DIR:= $(shell dirname $(MAIN_APP_FILE))
PROJECT_NAME=my-file-service

## ----------------------------------------------------------------------
## 		A little manual for using this Makefile.
## ----------------------------------------------------------------------


.PHONY: build
build:	swagger ## Compile the code into an executable application
	go build -v -o ./bin/main ${MAIN_APP_FILE}


.PHONY: docker-build
docker-build:	## Build docker image
	docker-compose build ${PROJECT_NAME}


.PHONY: run
run:	## Run application
	go run ${MAIN_APP_FILE}


MOCKS_DESTINATION=mocks
.PHONY: mocks
mocks: ## Generate mocks
	@echo "Generating mocks..."
	go generate ./...

.PHONY: test
test: mocks ## Run golang tests
	go test --short -race  -coverprofile=coverage.out -cover `go list ./... | grep -v mocks `

.PHONY: test.integration
test.integration: ## Run golang integration tests with dockerized environment
	go test -v ./tests

.PHONY: linter
linter:	## Run linter for *.go files
	revive -config .linter.toml  -exclude ./vendor/... -formatter unix ./...


.PHONY: docker-compose-up
docker-compose-up:	## Run application and app environment in docker
	docker-compose --env-file ${ENV_FILE} up

.PHONY: docker-compose-up-dev
docker-compose-up-dev:	## Run develop environment in docker
	docker-compose --env-file ${ENV_FILE} up minio



.PHONY: swagger
swagger:	## Generate swagger api specs
	swag init --output ./api --dir ${MAIN_APP_DIR},./internal/server --parseInternal true


.PHONY: help
help:     ## Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)


.DEFAULT_GOAL := build
