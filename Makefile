APP_VERSION = $(shell git describe --abbrev=0 --tags)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date -u "+%Y%m%d-%H%M")
VERSION_PKG = github.com/InjectiveLabs/chainlink-injective/version
IMAGE_NAME := gcr.io/injective-core/chainlink-injective

all:

image:
	docker build --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local -f Dockerfile .
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push:
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):latest

install: export GOPROXY=direct
install: export VERSION_FLAGS="-X $(VERSION_PKG).GitCommit=$(GIT_COMMIT) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)"
install:
	go install \
		-ldflags $(VERSION_FLAGS) \
		./cmd/...

.PHONY: install image push test gen

test: export GOPROXY=direct
test:
	go install github.com/onsi/ginkgo/ginkgo@latest
	ginkgo -v -r test/

mongo:
	mkdir -p var/mongo
	mongod --dbpath ./var/mongo --port 27017 --bind_ip 127.0.0.1 & echo $$! > var/mongo/mongod.pid;
	wait $(shell cat var/mongo/mongod.pid)

mongo-stop: var/mongo/mongod.pid
	kill `cat $<` && rm $<
