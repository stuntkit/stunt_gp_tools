
OUTPUT_DIR=bin/
COVER_FILE=cover.out

.PHONY: build test clean all

all: build test vet cover

build_all: build_linux_64 build_windows_32

build:
	go build  -ldflags "-s -w" -o ${OUTPUT_DIR} ./cmd/...

build_linux_64:
	GOOS=linux GOARCH=AMD64 go build  -ldflags "-s -w" -o ${OUTPUT_DIR} ./cmd/...

build_windows_32:
	GOOS=windows GOARCH=386 go build  -ldflags "-s -w" -o ${OUTPUT_DIR} ./cmd/...
test:
	go test ./cmd/... ./pkg/...

clean:
	rm -rf ${COVER_FILE}
	rm -rf ${OUTPUT_DIR}



cover:
	go test -coverprofile ${COVER_FILE} ./...

cover-show: cover
	go tool cover -html=${COVER_FILE}

vet:
	go vet ./cmd/... ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...	
