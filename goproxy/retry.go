package goproxy

import (
    log "github.com/Sirupsen/logrus"
    "sync"
    //"fmt"
    "github.com/tpjg/goriakpbc"
)

type RetryTask struct {
    bucket          *riak.Bucket
    key             string
    data            []byte
    retryTimes      int   
}

const (
    RETRY_WORKER_SIZE = 1
    RETRY_TASK_SIZE = 10000
)

var retryTaskWaitGroup *sync.WaitGroup
var retryTasks chan *RetryTask

func StartRetryWorker() {
    retryTaskWaitGroup = &sync.WaitGroup{}
    retryTasks = make(chan *RetryTask, RETRY_TASK_SIZE)

    for i:= 0; i < RETRY_WORKER_SIZE; i++{
        go func() {
            for {
                task,ok := <- retryTasks
                if !ok {
                    log.Error("get retry tasks from channel failed")
                }
                task.Do()
            }
        }()
    }
    log.Debug("start retry task done")
}

func NewRetryTask(bucket *riak.Bucket, key string, data []byte) *RetryTask {
    task := new(RetryTask)
    task.bucket = bucket
    task.key = key
    task.data = data
    task.retryTimes = 1

    retryTaskWaitGroup.Add(1)

    return task
}

func WaitRetryTaskDone() {
    log.Debug("wait retry task done..., retry_tasks:", len(retryTasks))
    retryTaskWaitGroup.Wait()
    log.Debug("retry task done, retry_tasks:", len(retryTasks))
}

func (task *RetryTask) Enqueue() {
    retryTasks <- task
}

func (task *RetryTask) Do() {
    err := DoRiakSet(task.bucket, task.key, task.data)

    if err != nil {
        if task.retryTimes < 3 {
            log.WithFields(log.Fields{
                "bucket": task.bucket.Name(),
                "key": task.key,
                "data": string(task.data),
                "retry_times": task.retryTimes,
            }).Debug("riak set failed:" + err.Error())

            task.retryTimes += 1
            task.Enqueue()
        } else {
            log.WithFields(log.Fields{
                "bucket": task.bucket.Name(),
                "key": task.key,
                "data": string(task.data),
                "retry_times": task.retryTimes,
            }).Error("riak set failed after retry:" + err.Error())
        }
    }

    retryTaskWaitGroup.Done()
}
