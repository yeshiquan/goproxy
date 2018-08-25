package goproxy

import (
    log "github.com/Sirupsen/logrus"
    "errors"
)

var ErrClosed = errors.New("pool is closed")

type ConnPool struct {
    conns       chan *Conn
}

func NewConnPool(maxCap int) (*ConnPool, error) {
    if maxCap <= 0 {
        return nil, errors.New("invalid capacity settings")
    }

    pool := &ConnPool{
        conns:          make(chan *Conn, maxCap),
    }
    
    for i := 0; i < maxCap; i++ {
        conn, err := NewConn()
        if err != nil {
            log.Error("create conn failed")
            return nil,err
        }
        conn.pool = pool
        pool.conns <- conn
    }

    return pool, nil
}

func (p *ConnPool) Get() (*Conn, error) {
    if p.conns == nil {
        return nil, ErrClosed
    }
    conn := <- p.conns

    return conn, nil
}

func (p *ConnPool) Put(conn *Conn) error {
    if conn == nil {
        return errors.New("connection is nil. rejecting")
    }
    (*conn.conn).Close()

    if p.conns == nil {
        return nil
    }

    select {
    case p.conns <- conn:
        return nil
    default:
        //pool is full
        return nil
    }
}
