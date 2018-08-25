package goproxy

import (
    "net"
    "fmt"
    "time"
    "bytes"
    "errors"
    "sync"
    "io"
    "../nshead"
)

type Conn struct {
    mu              sync.Mutex
    conn            *net.Conn
    pool            *ConnPool
    bodyBuf         []byte
    headBuf         []byte
    hd              *nshead.NsHead
    sendBuf         *bytes.Buffer 
    tmpBuf          []byte
}

func NewConn() (c *Conn, err error) {
    c = &Conn{}
    c.hd = nshead.NewNsHead()
    headerLen,bodyLen,tmpLen,sendLen := int(nshead.NSHEAD_HEADER_LEN),4096,16,4096
    buf := make([]byte, headerLen + bodyLen + tmpLen + sendLen)

    start := 0 
    c.headBuf = buf[start:start+headerLen]
    start += headerLen
    c.bodyBuf = buf[start: start + bodyLen]
    start += bodyLen
    c.tmpBuf = buf[start: start + tmpLen]
    start += tmpLen
    c.sendBuf = bytes.NewBuffer(buf[start: start+sendLen])

    return c, nil
}

func (c *Conn) Read() error {
    conn := *c.conn

    conn.SetReadDeadline(time.Now().Add(5 * time.Second))  
    _, err := io.ReadFull(conn, c.headBuf[0:nshead.NSHEAD_HEADER_LEN])
    if err != nil {
        return err
    }
    err = c.hd.Decode(c.headBuf)
    if err != nil {
        return err
    }

    if c.hd.BodyLen > 1000000 {
        return errors.New(fmt.Sprintf("request body too large: %d", c.hd.BodyLen))
    }
   
    c.bodyBuf = GrowSlice(c.bodyBuf, c.hd.BodyLen) 
    _, err = io.ReadFull(conn, c.bodyBuf[0:c.hd.BodyLen])
    if err != nil {
        return err
    }

    return nil
}

func (c *Conn) Write(data []byte) (int, error) {
    conn := *c.conn
    c.hd.BodyLen = uint32(len(data))
    c.hd.Encode(c.sendBuf, c.tmpBuf)
    c.sendBuf.Write(data)
    n, err := c.sendBuf.WriteTo(conn)

    return int(n), err
}

func GrowSlice(buf []byte, size uint32) []byte {
    for uint32(cap(buf)) < size {
        buf = append(buf, 0)
    }

    return buf[0:size]
}
