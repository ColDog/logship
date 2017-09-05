export VERSION=`cat version`
export REPO=coldog/logship

install:
	go install github.com/$(REPO)
.PHONY: install

run: install
	logship
.PHONY: run

test:
	go test -race -v ./...
.PHONY: test

cov:
	go test -cover ./...
.PHONY: test

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o logship .
	docker build -t $(REPO):$(VERSION) .
	docker build -t $(REPO):$(VERSION)-builder -f Dockerfile.builder .
	rm logship
.PHONY: build

push:
	docker push $(REPO):$(VERSION)
.PHONY: push
