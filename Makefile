.PHONY: all clean

all: build

EXE = $(GOPATH)/bin/deadman
SRC = $(shell find ./ -type f -name '*.go')

$(EXE): $(SRC)
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-s -w -static"' -o $(EXE) .

build: $(EXE)

.PHONY: dep
dep:
ifeq ($(shell command -v dep 2> /dev/null),)
	go get -u -v github.com/golang/dep/cmd/dep
endif

.PHONY: deps
deps: dep
	dep ensure -v

docker: deps
	docker build -t deadman .

clean:
	rm -f $(EXE)
