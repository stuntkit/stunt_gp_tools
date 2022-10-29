
OUTPUT_DIR=bin/
COVER_FILE=cover.out

.PHONY: build test clean all

all: build test vet cover

build:
	go build  -ldflags "-s -w" -o ${OUTPUT_DIR} ./cmd/...

build_windows:
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
