package goproxy

import (
    "os"
    "os/signal"
    "syscall"
    log "github.com/Sirupsen/logrus"
    "fmt"
    "net"
    "time"
)

const (
    CONN_TYPE = "tcp"
    RESPONSED = 1
    CLOSED = 2
)

type Server struct {
    pool        *ConnPool
    socket      *net.TCPListener
}

func HandleResponse(c *Conn, data []byte) {
    c.Write(data)
}

func (s *Server)HandleRequest(conn *net.Conn) {
     c,_ := s.pool.Get() 
     c.conn = conn
     go Process(c)
}

func (s *Server) Start() {
    var err error
    SetupConfig()
    if os.Getenv("GOPROXY_GRACEFUL_RESTART") == "true" {
        var fd uintptr = 3
        file := os.NewFile(fd, "/tmp/sock-go-graceful-restart")
        listener, err := net.FileListener(file)
        if err != nil {
            fmt.Println("File to recover socket from file descriptor: " + err.Error())
        }
        listenerTCP, ok := listener.(*net.TCPListener)
        if !ok {
            fmt.Println(fmt.Sprintf("File descriptor %d is not a valid TCP socket", fd))
        }
        s.socket = listenerTCP
    } else {
        addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", CONFIG.ListenPort))
        if err != nil {
            fmt.Println("fail to resolve addr: " +  err.Error())
            return
        }
        socket, err := net.ListenTCP("tcp", addr)
        if err != nil {
            fmt.Println("fail to listen: " +  err.Error())
            return
        }
        s.socket = socket
    }

    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer s.socket.Close()

    s.SetupLogger()
    s.pool,err = NewConnPool(CONFIG.ConnPoolSize)
    if err != nil {
        fmt.Println("Error create connection pool:", err.Error())
        os.Exit(1)
    }

    err = CreateRiakConnectPool()
    if err != nil {
        log.Error("create riak connect pool failed: " + err.Error())
        os.Exit(1)
    }
    go PingRiak()

    go StartAsyncWorker()
    go StartRetryWorker()
    go s.StartAcceptLoop()
    go StartHttpServer()

    fmt.Printf("Server start at -- port:%v\n", CONFIG.ListenPort)

    s.SetupSignal()
}

func (s *Server) StartAcceptLoop() {
    for {
        conn, err := s.socket.Accept()
        if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            log.Info("stop accepting connections")
            return
        } 
        go s.HandleRequest(&conn)
    }
}

func (s *Server) Stop() {
    // Accept will instantly return a timeout error
    s.socket.SetDeadline(time.Now())
}

func (s *Server) ListenerFD() (uintptr, error) {
    file, err := s.socket.File()
    if err != nil {
        return 0, err
    }

    return file.Fd(), nil
}

func (s *Server) Wait() {
    WaitCmdDone()
    WaitRetryTaskDone()
}

func (s *Server) SetupLogger() {
    f, err := os.OpenFile(CONFIG.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("error opening file", CONFIG.LogFile, err)
        os.Exit(1)
    }
    log.SetOutput(f)
    log.SetFormatter(&log.TextFormatter{
        FullTimestamp :     true,
        TimestampFormat:    "2006-01-02 15:04:05",//时间格式奇葩
        DisableColors:      true,
    })
    if CONFIG.LogLevel == "Debug" {
        log.SetLevel(log.DebugLevel)
    } else if CONFIG.LogLevel == "Info" {
        log.SetLevel(log.InfoLevel)
    } else {
        log.SetLevel(log.ErrorLevel)
    } 
}

func (s *Server) SetupSignal() {
    signals := make(chan os.Signal)
    signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
    for sig := range signals {
        if sig == syscall.SIGTERM || sig == syscall.SIGINT {
            // Stop accepting new connections
            log.Info("server shutdown,wait connection finish...")
            s.Stop()
            // Wait for existing connections to finish
            s.Wait()
            // Then the program exists
            log.Info("server shutdown successful")
            os.Exit(0)
        } else if sig == syscall.SIGHUP {
            // Stop accepting requests
            s.Stop()
            // Get socket file descriptor to pass it to fork
            listenerFD, err := s.ListenerFD()
            if err != nil {
                log.Fatal("Fail to get socket file descriptor:", err)
            }
            // Set a flag for the new process start process
            os.Setenv("GOPROXY_GRACEFUL_RESTART", "true")
            execSpec := &syscall.ProcAttr{
                Env:   os.Environ(),
                Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFD},
            }
            // Fork exec the new version of your server
            fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
            if err != nil {
                log.Fatal("Fail to fork", err)
            }
            log.Println("SIGHUP received: fork-exec to", fork)
            // Wait for all conections to be finished
            s.Wait()
            log.Println(os.Getpid(), "Server gracefully shutdown")

            // Stop the old server, all the connections 
            // have been closed and the new one is running
            os.Exit(0)
        } 
    }
}
