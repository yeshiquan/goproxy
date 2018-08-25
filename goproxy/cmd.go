package goproxy

import (
    "errors"
    log "github.com/Sirupsen/logrus"
    "sync"
    //"fmt"
    "reflect"
    "encoding/json"
    "github.com/tpjg/goriakpbc"
)

var (
    ErrInvalidParameter = errors.New("invalid parameter")
    ErrUnsupportMethod = errors.New("unsupport method")
)

const (
    ASYNC_WORKER_SIZE = 24
    WAIT_TASK_SIZE = 100000
)

type Cmd struct {
    c       *Conn
    bucket  *riak.Bucket
    req     map[string]interface{}
    result  *Result
    method  string
    async   bool
}

var waitTasks chan *Cmd
var cmdWaitGroup *sync.WaitGroup

func StartAsyncWorker() {
    waitTasks = make(chan *Cmd, WAIT_TASK_SIZE)
    cmdWaitGroup = &sync.WaitGroup{}

    for i:= 0; i < ASYNC_WORKER_SIZE; i++{
        go func() {
            for {
                cmd,ok := <- waitTasks
                if !ok {
                    log.Error("get async tasks from channel failed")
                }
                cmd.Do()
            }
        }()
    }
    log.Debug("start async task done")
}

func WaitCmdDone() {
    log.Debug("wait async task done..., async_tasks:", len(waitTasks))
    cmdWaitGroup.Wait()
    log.Debug("async task done, async_tasks:", len(waitTasks))
}

func NewCmd(c *Conn, bucket *riak.Bucket, req map[string]interface{}, result *Result, method string) *Cmd {
    cmd := new(Cmd)
    cmd.c = c
    cmd.bucket = bucket
    cmd.result = result
    cmd.method = method
    cmd.req = req

    cmdWaitGroup.Add(1)

    return cmd
}


func (cmd *Cmd) Enqueue() {
    waitTasks <- cmd 
}

func (cmd *Cmd) DoGet() error {
    req := cmd.req
    key,ok := req["params"].(string)
    if !ok || key == ""{
        return ErrInvalidParameter
    }

    data, err := DoRiakGet(cmd.bucket, key)

    if err != nil {
        return err
    }

    var dat interface{}
    if err := json.Unmarshal(data, &dat); err != nil {
        return err
    }
    cmd.result.Data = dat

    return nil
}

func (cmd *Cmd) DoMget() error {
    req := cmd.req
    keys,ok := req["params"].([]interface{})
    if !ok {
        return ErrInvalidParameter
    }
    var result = struct{
        sync.RWMutex
        m map[string]interface{}
    }{m: make(map[string]interface{})}

    length := len(keys)
    var w sync.WaitGroup
    w.Add(length)

    for i:= 0; i < length; i++ {
        key,ok := keys[i].(string)
        if !ok {
            w.Done()
            continue
        }
        
        go func() {
            data, err := DoRiakGet(cmd.bucket, key)
            if err != nil {
                w.Done()
                return 
            }

            var item interface{}
            if err := json.Unmarshal(data, &item); err != nil {
                w.Done()
                return
            }
            result.Lock()
            result.m[key] = item
            result.Unlock()
            w.Done()
        }()
    }
    w.Wait()
    v := reflect.ValueOf(result.m)
    cmd.result.Data = v.Interface()
    return nil
}

func (cmd *Cmd) DoSet() error {
    req := cmd.req
    params,ok := req["params"].(map[string]interface{})

    if !ok {
        return ErrInvalidParameter
    } 

    for key, v := range params {
        switch value := v.(type) {
        default:
            jsonStr,_ := json.Marshal(value)
            return DoRiakSet(cmd.bucket, key, []byte(jsonStr))
        }
    }            

    return nil
}

func (cmd *Cmd) DoMset() error {
    req := cmd.req
    params,ok := req["params"].(map[string]interface{})

    if !ok {
        return ErrInvalidParameter
    } 

    var w sync.WaitGroup
    w.Add(len(params))
    for key, v := range params {
        key := key  //必须重新创建变量，因为多个goutine共享了for循环中的key
        switch value := v.(type) {
        default:
            go func() {
                jsonStr,_ := json.Marshal(value)
                err := DoRiakSet(cmd.bucket, key, []byte(jsonStr))
                if err != nil {
                    task := NewRetryTask(cmd.bucket, key, []byte(jsonStr))
                    task.Enqueue()
                    log.WithFields(log.Fields{
                        "key": key,
                    }).Debug("do mset failed")
                }
                w.Done()
            }()
        }
    }
    w.Wait()

    return nil
}

func (cmd *Cmd) Do() {
    cmdMap := map[string]func() error {
        "get": cmd.DoGet,
        "mget": cmd.DoMget,
        "set": cmd.DoSet,
        "mset": cmd.DoMset,
    }

    var err error
    callFunc,ok := cmdMap[cmd.method]
    if ok {
        err = callFunc()
    } else {
        err = ErrUnsupportMethod 
    }
    if !cmd.async {
        if err != nil {
            SendErrorResponse(cmd.c, cmd.result, err)
        } else {
            SendResponse(cmd.c, cmd.result)
        }
    }

    cmdWaitGroup.Done()
}
