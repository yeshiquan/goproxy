package goproxy 

import (
    "os"
    "encoding/json"
    "path/filepath"
    "fmt"
)

type Config struct {
    ListenPort      int     `json:"listen_port"`
    HttpPort        int     `json:"http_port"`
    LogFile         string  `json:"log_file"`
    ConnPoolSize    int     `json:"conn_pool_size"`
    LogLevel        string  `json:"log_level"`
    RiakPoolSize    int     `json:"riak_pool_size"`
    RiakAddress     string  `json:"riak_address"`
}

var CONFIG Config

func SetupConfig() {
    var filename string
    env := os.Getenv("DATA_PLATFORM_ENV")

    if env == "dev" {
        filename = "config/dev.json"
    } else if env == "test" {
        filename = "config/test.json"
    } else if env == "product" {
        filename = "config/product.json"
    } else {
        fmt.Println("export DATA_PLATFORM_ENV=dev/test/product first")
        os.Exit(1)
    }

    dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
    file, err := os.Open(dir + "/" + filename)
    if err != nil {
        file, err = os.Open(dir + "/../" + filename)
    }
    if err != nil {
        fmt.Println("open file failed: ", err)
        os.Exit(1)
    }
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&CONFIG)

    if err != nil {
        fmt.Println("can't parse config file:", err)
        os.Exit(1)
    }

    file.Close()
}
