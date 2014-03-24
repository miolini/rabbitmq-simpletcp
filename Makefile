all: build

build:
	go build -o rabbitmq-simpletcp rabbitmq-simpletcp.go
clean:
	rm -rf rabbitmq-simpletcp
depends:
	go get -d