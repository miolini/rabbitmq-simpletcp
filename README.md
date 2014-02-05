# RabbitMQ-SimpleTCP

This is daemon writen on Go language for receiver over TCP text stream CRLF separated and 
publish each line as a message in RabbitMQ exchange.

## Install

'''
# git clone https://github.com/miolini/rabbitmq-simpletcp
# make depends
# make
'''

## Usage

'''
# ./rabbitmq-simpletcp --help
Usage of ./rabbitmq-simpletcp:
  -addr="localhost:12010": tcp listen addr host:port
  -exchange="x-simpletcp": send all messages to this rabbitmq exchange
  -uri="amqp://guest:guest@localhost:5672/": rabbitmq server amqp uri
'''