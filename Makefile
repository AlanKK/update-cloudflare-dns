GOOS=darwin
GOARCH=amd64
BINARY_NAME=ip_update

all: darwin linux windows

darwin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) main.go

linux:
	GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY_NAME).linux main.go

windows:
	GOOS=windows GOARCH=$(GOARCH) go build -o $(BINARY_NAME).exe main.go

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).linux $(BINARY_NAME).exe