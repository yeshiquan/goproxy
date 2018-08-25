/**
   * @author yeshiquan(yeshiquan@baidu.com)
   * @date 2015-04-19
   * 对nshead的操作的GO版本
   *
   * nshead C结构体:
   * typedef struct _nshead_t
   * {
   *   unsigned short id;
   *   unsigned short version;
   *   unsigned int   log_id;
   *   char           provider[16];
   *   unsigned int   magic_num;
   *   unsigned int   reserved;
   *   unsigned int   body_len;
   * } nshead_t;
   *
   **/

package nshead

import (
    "bytes"
    "encoding/binary"
)

var DefaultMagicNum uint32 = 0xfb709394
var NSHEAD_HEADER_LEN uint32 = 36

type NsHead struct {
    Id       uint16
    Version  uint16
    LogId    uint32
    Provider string // max length 16
    MagicNum uint32
    Reserved uint32
    BodyLen  uint32
}

func NewNsHead() *NsHead {
    return &NsHead{Id:0,Version: 0,LogId:0,Provider:"",MagicNum: DefaultMagicNum,Reserved:0,BodyLen:0} 
}

func (nh *NsHead) Encode(b *bytes.Buffer, tmp []byte) []byte {
    b.Reset()
    binary.Write(b, binary.LittleEndian, nh.Id)
    binary.Write(b, binary.LittleEndian, nh.Version)
    binary.Write(b, binary.LittleEndian, nh.LogId)
    tmp = tmp[0:16]
    copy(tmp, []byte(nh.Provider))
    binary.Write(b, binary.LittleEndian, tmp)
    binary.Write(b, binary.LittleEndian, nh.MagicNum)
    binary.Write(b, binary.LittleEndian, nh.Reserved)
    binary.Write(b, binary.LittleEndian, nh.BodyLen)

    return b.Bytes()
}

func (nh *NsHead) Decode(buf []byte) error {
    b := bytes.NewBuffer(buf)
    binary.Read(b, binary.LittleEndian, &nh.Id)
    binary.Read(b, binary.LittleEndian, &nh.Version)
    binary.Read(b, binary.LittleEndian, &nh.LogId)
    nh.Provider = string(buf[8:24])
    b.Next(16)
    binary.Read(b, binary.LittleEndian, &nh.MagicNum)
    binary.Read(b, binary.LittleEndian, &nh.Reserved)
    binary.Read(b, binary.LittleEndian, &nh.BodyLen)

    return nil
}
