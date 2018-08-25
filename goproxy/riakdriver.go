package goproxy

import (
    "github.com/tpjg/goriakpbc"
    //"fmt"
    //"errors"
    log "github.com/Sirupsen/logrus"
    "os"
    "sync"
    "time"
)

var bucketMap = struct{
    sync.RWMutex
    m map[string]*riak.Bucket
}{m: make(map[string]*riak.Bucket)}

func CreateRiakConnectPool() error {
    return riak.ConnectClientPool(CONFIG.RiakAddress, CONFIG.RiakPoolSize)
}

func PingRiak() {
    for {
        err := riak.Ping()
        if err != nil {
            log.Error("riak server down: ", err)
            os.Exit(1)
        }
        time.Sleep(1 * time.Second)
    }
}

func GetBucket(bucket_type string, bucket_name string) (*riak.Bucket, error) {
    key := bucket_type + "##" + bucket_name
    var err error 
    //maps are not safe for concurrent use in golang
    bucketMap.RLock()
    bucket,ok := bucketMap.m[key]
    bucketMap.RUnlock()

    if !ok{
        bucket,err = riak.NewBucketType(bucket_type,bucket_name)
        if err == nil {
            bucketMap.Lock()
            bucketMap.m[key] = bucket
            bucketMap.Unlock()
        }
        return bucket,err
    }

    return bucket, nil
}

func DoRiakSet(bucket *riak.Bucket, key string, data []byte) error {
    startTime := time.Now().UTC()

    obj := bucket.NewObject(key)
    obj.ContentType = "application/json"
    obj.Data = data
    err := obj.Store()
    if err != nil {
        log.WithFields(log.Fields{
            "bucket": bucket.Name(),
            "key": key,
            "data": string(data),
        }).Error("riak set failed:" + err.Error())
        return err
    } 

    endTime := time.Now().UTC()
    var duration time.Duration = endTime.Sub(startTime)

    log.WithFields(log.Fields{
        "bucket": bucket.Name(),
        "key": key,
        "elapsed": duration,
    }).Debug("riak set done")

    return nil
}

func DoRiakGet(bucket *riak.Bucket, key string) ([]byte, error) {
    startTime := time.Now().UTC()
    var data []byte

    obj, err := bucket.Get(key)
    if err != nil {
		if err.Error() != "Object not found" {
        	log.WithFields(log.Fields{
        	    "bucket": bucket.Name(),
        	    "key": key,
        	}).Error("riak get failed:" + err.Error())
		} else {
        	log.WithFields(log.Fields{
        	    "bucket": bucket.Name(),
        	    "key": key,
        	}).Info("riak get failed:" + err.Error())
		}
    	return nil,err
    }

    if obj.Conflict() {
        log.WithFields(log.Fields{
            "bucket": bucket.Name(),
            "key": key,
        }).Warning("riak get data multiple values or siblings")
        //从多个值中任取一个，这里取最后一个，拍脑袋想的
        data = obj.Siblings[len(obj.Siblings)-1].Data
        //同时解决冲突
        obj.Reload()
    } else {
        data = obj.Data
    }

    endTime := time.Now().UTC()
    var duration time.Duration = endTime.Sub(startTime)

    log.WithFields(log.Fields{
        "bucket": bucket.Name(),
        "key": key,
        "elapsed": duration,
    }).Debug("riak get done")

    return data,err
}
