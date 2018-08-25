package main

import (
    "runtime"
    "./goproxy"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU()) 
    server := &goproxy.Server{}
    server.Start()
}
