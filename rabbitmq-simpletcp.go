package main

import "log"
import "flag"
import "net"
import "bufio"
import "time"
import "github.com/streadway/amqp"

const (
    RECV_BUF_LEN = 2048
)

func listenTCP(addr string, dataChan chan string) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        log.Printf("error: %s", err)
        return
    }
    for {
        conn, err := listener.Accept()
        if err != nil {
            println("error: %s", err)
            return
        }
        log.Printf("new connection")
        go func(conn net.Conn) {
            defer conn.Close()
            scanner := bufio.NewScanner(conn)
            for scanner.Scan() {
                line := scanner.Text()
                dataChan <- line
            }
        }(conn)
    }
}

func startPublisher(uri string, exchange string, dataChan chan string) {
    queueConn, err := amqp.Dial(uri)
    if err != nil {
        log.Printf("error: %s", err)
        return
    }
    defer queueConn.Close()
    queueChan, err := queueConn.Channel()
    if err != nil {
        log.Printf("error: %s", err)
        return
    }
    defer queueChan.Close()
    for {
        select {
            case data := <- dataChan:
                msg := amqp.Publishing{ContentType:"text/plain",Body: []byte(data)}
                err := queueChan.Publish(exchange, "", false, false, msg)
                if err != nil {
                    log.Printf("error: %s", err)
                    break
                }
        }
    }
}

func main() {
    log.Printf("rabbitmq-simpletcp transmitter")

    var (
        uri string
        exchange string
        addr string
    )

    flag.StringVar(&addr, "addr", "localhost:12010", "tcp listen addr host:port")
    flag.StringVar(&uri, "uri", "amqp://guest:guest@localhost:5672/", "rabbitmq server amqp uri")
    flag.StringVar(&exchange, "exchange", "x-simpletcp", "send all messages to this rabbitmq exchange")
    flag.Parse()

    log.Printf("addr:      %s", addr)
    log.Printf("uri:       %s", uri)
    log.Printf("exchange:  %s", exchange)

    dataChan := make(chan string, 1000)

    go func() {
        for {
            log.Printf("start tcp listener")
            listenTCP(addr, dataChan)
            time.Sleep(time.Second)
        }
    }()

    go func() {
        for {
            log.Printf("start amqp publisher")
            startPublisher(uri, exchange, dataChan)
            time.Sleep(time.Second)
        }
    }()

    <- make(chan bool)
}