GOOS=darwin
GOARCH=amd64
BINARY_NAME=ip_update

all: darwin linux windows

darwin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) main.go pushbullet.go

linux:
	GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY_NAME).linux main.go pushbullet.go

windows:
	GOOS=windows GOARCH=$(GOARCH) go build -o $(BINARY_NAME).exe main.go pushbullet.go

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).linux $(BINARY_NAME).exe