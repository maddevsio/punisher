TARGET=punisher

all: clean fmt build

clean:
	rm -rf $(TARGET)

test:
	go test ./...

build:
	go build -o $(TARGET) main.go

fmt:
	go fmt ./...

migrate:
	goose -dir migrations mysql "root:root@tcp(localhost:3306)/interns"  up

build_linux:
	GOOS=linux GOARCH=amd64 go build -o $(TARGET) main.go

build_docker:
	docker build -t punisher .

docker: build_linux build_docker