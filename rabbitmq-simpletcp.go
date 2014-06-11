package main

import "log"
import "flag"
import "net"
import "io"
import "bufio"
import "time"
import "github.com/streadway/amqp"
import "github.com/miolini/godaemon"
import "compress/gzip"

const (
    RECV_BUF_LEN = 2048
)

func readClient(reader io.Reader, dataChan chan string, compress bool) {
    var scanner *bufio.Scanner
    if compress {
        gzipReader, err := gzip.NewReader(reader)
        if err != nil {
            log.Printf("gzip error: %s", err)
            return
        }
        scanner = bufio.NewScanner(gzipReader)
    } else {
        scanner = bufio.NewScanner(reader)
    }
    for scanner.Scan() {
        dataChan <- scanner.Text()
    }
}

func listenTCP(addr string, dataChan chan string, compress bool) {
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
        go readClient(conn, dataChan, compress)
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
                    return
                }
        }
    }
}

func startTcpSender(addr string, copyChan chan string) {
    client, err := net.Dial("tcp", addr)
    if err != nil {
        log.Printf("error: %s", err)
        return
    }
    defer client.Close()
    for {
        select {
        case data := <- copyChan:
            _, err := io.WriteString(client, data + "\r\n")
            if err != nil {
                log.Printf("can't write to %s: %s", addr, err)
                return
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
        gzip bool
        pidFile string
    )

    flag.StringVar(&addr, "addr", "localhost:12010", "tcp listen addr host:port")
    flag.StringVar(&uri, "uri", "amqp://guest:guest@localhost:5672/", "rabbitmq server amqp uri")
    flag.StringVar(&exchange, "exchange", "x-simpletcp", "send all messages to this rabbitmq exchange")
    flag.BoolVar(&gzip, "gzip" , false, "enable gzip decompress")
    flag.StringVar(&pidFile, "pidfile", "", "path to pid file")
    flag.Parse()

    log.Printf("addr:      %s", addr)
    log.Printf("uri:       %s", uri)
    log.Printf("exchange:  %s", exchange)
    log.Printf("pidfile:   %s", pidFile)

    godaemon.WritePidFile(pidFile)

    dataChan := make(chan string, 1000)

    go func() {
        for {
            log.Printf("start tcp listener")
            listenTCP(addr, dataChan, gzip)
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
