package main

import (
    "net"
    "os"
    "io/ioutil"
    "time"
    "fmt"
    "sync"
    "strconv"
)

const (
    DATA_FILE = "/home/yeshiquan/echo.dat"
    SERVER_ADDR = "10.95.31.39:8002" //baidu-rpc
    //SERVER_ADDR = "10.95.31.39:8033" //golang
    //SERVER_ADDR = "10.16.82.183:8877"
)

var data []byte
var WaitGroup *sync.WaitGroup = &sync.WaitGroup{}

func readFile() {
    var err error
    data, err = ioutil.ReadFile(DATA_FILE)
    if err != nil {
        panic(err)
    }
}


var servAddr string 
var tcpAddr *net.TCPAddr

func initSocket() {
    var err error
    tcpAddr, err = net.ResolveTCPAddr("tcp", SERVER_ADDR)
    if err != nil {
        println("ResolveTCPAddr failed:", err.Error())
        os.Exit(1)
    }
}

func request(th int, n int) {
    reply := make([]byte, 18192)
    for i := 0; i < n; i++ {
        conn, err := net.DialTCP("tcp", nil, tcpAddr)
        if err != nil {
            println("Dial failed:", err.Error())
            os.Exit(1)
        }

        _, err = conn.Write(data)
        if err != nil {
            println("Write to server failed:", err.Error())
            os.Exit(1)
        }
     
        _, err = conn.Read(reply)
        if err != nil {
            println("Write to server failed:", err.Error())
            os.Exit(1)
        }
        conn.Close()
    }
    WaitGroup.Done()
    fmt.Printf("thread %d finished\n", th)
}

func main() {
    readFile()
    initSocket()
    c,_ := strconv.Atoi(os.Args[1])
    n,_ := strconv.Atoi(os.Args[2])
    WaitGroup.Add(c)
    startTime := time.Now().UTC()
    for i:= 0; i < c; i++ {
        go request(i+1, n)
    }
    WaitGroup.Wait()
    endTime := time.Now().UTC()
    var cost time.Duration = endTime.Sub(startTime)
    seconds := cost.Seconds()

    fmt.Printf("Total Time: %v\n", cost)
    fmt.Printf("QPS: %v\n", (float64(c*n)/seconds))
    fmt.Printf("Time per request: %v ms\n", (seconds*1000.0/float64(c*n)))
}
