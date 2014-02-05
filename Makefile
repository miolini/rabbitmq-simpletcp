all:
	go build -o rabbitmq-simpletcp rabbitmq-simpletcp.go

depends:
	go get github.com/streadway/amqp