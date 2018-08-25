package goproxy

import (
    "fmt"
    "io"
    "runtime"
    "encoding/json"
    "net/http"
    "runtime/pprof"
    "github.com/gorilla/mux"
    log "github.com/Sirupsen/logrus"
)
import _ "net/http/pprof"

type ShortUrl struct {
    Type    int     `json:"type"`
    Url     string  `json:"url"`
}

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    token := vars["token"]
    if token != "1q2w3e" {
        return
    }
    navStr := `<a href="/">Back</a></br>`
    w.Header().Set("Content-Type", "text/html")
    p := pprof.Lookup("goroutine")
    io.WriteString(w, navStr)
    io.WriteString(w, "<pre>")
    p.WriteTo(w, 1)
    io.WriteString(w, "</pre>")
}

func heapmemHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    token := vars["token"]
    if token != "1q2w3e" {
        return
    }
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    navStr := `<a href="/">Back</a></br>`
    statStr := fmt.Sprintf("sys:\t\t%d KB\nalloc:\t\t%d KB\nidle:\t\t%d KB\nreleased:\t%d KB\n", 
            m.HeapSys/1024, m.HeapAlloc/1024, m.HeapIdle/1024, m.HeapReleased/1024)
    w.Header().Set("Content-Type", "text/html")
    io.WriteString(w, navStr)
    io.WriteString(w, "<pre>")
    io.WriteString(w, statStr)
    io.WriteString(w, "</pre>")
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    helpMsg := 
`
support command:<br/>
    <ul>
        <li><a href="/goroutine">goroutine</a></li>
        <li><a href="/heapmem">heapmem</a></li>
    </ul>
`
    io.WriteString(w, helpMsg)
}

func shortUrlHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    md5 := vars["md5"]
    product := vars["product"]
    bucketType := "short_addr"
    bucketName := product + "_short_addr"
    bucket,err := GetBucket(bucketType, bucketName)
    url := "http://www.baidu.com/search/error.html"

    log.WithFields(log.Fields{
            "md5": md5,
    }).Info("shorturl redirect")
    
    if err != nil {
        log.WithFields(log.Fields{
            "md5": md5,
            "bucket_name": bucketName,
            "err_msg": err.Error(),
        }).Error("shorturl get bucket failed")
        http.Redirect(w, r, url, 301)
        return
    }

    data, err := DoRiakGet(bucket, md5)

    if err != nil {
		if err.Error() == "Object not found" {
            log.WithFields(log.Fields{
                "md5": md5,
                "bucket_name": bucketName,
                "err_msg": err.Error(),
            }).Info("shorturl riak get failed")
        } else {
            log.WithFields(log.Fields{
                "md5": md5,
                "bucket_name": bucketName,
                "err_msg": err.Error(),
            }).Error("shorturl riak get failed")
        }
        http.Redirect(w, r, url, 301)
        return
    }

    var shortUrl ShortUrl
    if err := json.Unmarshal(data, &shortUrl); err != nil {
        log.WithFields(log.Fields{
            "md5": md5,
            "bucket_name": bucketName,
            "err_msg": err.Error(),
        }).Debug("shorturl riak get failed")
        http.Redirect(w, r, url, 301)
        return
    }

    http.Redirect(w, r, shortUrl.Url, 301)
}

func StartHttpServer() {
    r := mux.NewRouter()
    r.HandleFunc("/", helpHandler)
    r.HandleFunc("/goroutine/{token}", goroutineHandler)
    r.HandleFunc("/t/{product}/{md5}", shortUrlHandler)
    r.HandleFunc("/heapmem/{token}", heapmemHandler)
    http.Handle("/", r)
    fmt.Printf("Http Server start at -- port:%d\n", CONFIG.HttpPort)
    http.ListenAndServe(fmt.Sprintf(":%d", CONFIG.HttpPort), nil)
}
