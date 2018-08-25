package goproxy

import (
    log "github.com/Sirupsen/logrus"
    //"fmt"
    "time"
    "net"
    "io"
    "encoding/json"
)


type Status struct {
    Errno   int     `json:"errno"`
    Errmsg  string  `json:"errmsg"`
} 

type Result struct {
    Status   *Status            `json:"status"`
    Data     interface{}        `json:"data"`
}


func SendResponse(c *Conn, result *Result) {
    jsonStr, _ := json.Marshal(result)
    HandleResponse(c, []byte(jsonStr))
}

func SendErrorResponse(c *Conn, result *Result, err error) {
    result.Status.Errno = -1
    result.Status.Errmsg = err.Error()
    SendResponse(c, result)
}

func Process(c *Conn) {
    var req map[string]interface{}
    result := &Result{}
    result.Status = &Status{Errno: 0, Errmsg: ""}
    var bucketName string
    var bucketType string
    var method string

    for {
        result.Status.Errno = 0
        result.Status.Errmsg = ""
        result.Data = nil
        err := c.Read()

        startTime := time.Now().UTC()
        if err != nil {
            if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
                log.Info("read data from client timeout: " + err.Error())
            } else if err != io.EOF {
                log.Info("read data from client failed: " + err.Error())
            }   
            break
        }   

        //fmt.Printf(string(c.bodyBuf) + "\n")

        if err := json.Unmarshal([]byte(string(c.bodyBuf)), &req); err != nil {
            SendErrorResponse(c, result, err)
            continue
        }

        var okm,okb,okt bool
        method,okm = req["method"].(string)
        bucketName,okb = req["bucket"].(string)
        bucketType,okt = req["bucket_type"].(string)

        if (method == "" || bucketName == "" || ! okm || !okb) {
            SendErrorResponse(c, result, ErrInvalidParameter)
            continue
        }

        if bucketType == "" || !okt {
            bucketType = bucketName //bucketType默认和bucketName一样
        }
        bucket,err := GetBucket(bucketType,bucketName)
        if err != nil {
            SendErrorResponse(c, result, err)
            continue
        }

        //fmt.Printf("method: [%v]\n", method)
        //fmt.Printf("bucket: [%v]\n", bucketName)

        cmd := NewCmd(c, bucket, req, result, method);
        async,oka := req["async"].(bool)
        if method == "mset" {
            cmd.async = (async || !oka)
        } else {
            cmd.async = async
        }

        if cmd.async {
            SendResponse(cmd.c, cmd.result)
            cmd.Enqueue()
        } else {
            cmd.Do()
        }

        endTime := time.Now().UTC()
        var duration time.Duration = endTime.Sub(startTime)
        log.WithFields(log.Fields{
            "bucket": bucketName,
            "method": method,
            "async": cmd.async,
            "elapsed": duration,
        }).Info("process command done")
    }

    c.pool.Put(c)
}

