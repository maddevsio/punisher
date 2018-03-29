TARGET=comedian

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