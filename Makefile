BUILD_DIR 		:= build

.DEFAULT_GOAL := build

.PHONY: init
init:
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/modocache/gover
	go mod download

.PHONY: lint
lint:
	GO111MODULE=off $(GOPATH)/bin/gometalinter --disable-all --config .gometalinter.json ./...

.PHONY: build
build:
	env GOOS=linux go build -o $(BUILD_DIR)/k8s-rmq-autoscaler .

.PHONY: test
test:
	$(GOPATH)/bin/ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --progress --compilers=2

.PHONY: cover
cover:
	$(GOPATH)/bin/gover . coverage.txt

.PHONY: format
format:
	gofmt -w -s .
	goimports -w .
