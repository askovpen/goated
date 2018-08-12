package msgapi

import (
  "bufio"
  "encoding/binary"
  "bytes"
  "errors"
  "fmt"
  "github.com/askovpen/goated/lib/types"
  "github.com/askovpen/goated/lib/utils"
  "log"
  "os"
  "strings"
)

type Squish struct {
  AreaPath string
  AreaName string
  indexStructure []sqi_s
}

type sqi_s struct {
  Offset uint32
  MessageNum uint32
  CRC uint32
}

type sqd_h struct {
  Id, NextFrame, PrevFrame, FrameLength, MsgLength, CLen uint32
  FrameType, Rsvd uint16
  Attr  uint32
  From, To   [36]byte
  Subject [72]byte
  FromZone, FromNet, FromNode , FromPoint uint16
  ToZone, ToNet, ToNode, ToPoint uint16
  DateWritten, DateArrived, Utc uint16
  ReplyTo uint32
  Replies [9]uint32
  UMsgId  uint32
  Date  [20]byte
}

func (s *Squish) GetMsg(position uint32) (*Message, error) {
  if len(s.indexStructure)==0 { return nil, errors.New("Empty Area") }
  if position==0 { position=1 }
  f, err := os.Open(s.AreaPath+".sqd")
  if err!=nil {
    return nil, err
  }
  defer f.Close()
  f.Seek(int64(s.indexStructure[position-1].Offset),0)
  var header []byte
  header=make([]byte,266)
  f.Read(header)
  headerb:=bytes.NewBuffer(header)
  var sqdh sqd_h
  if err=utils.ReadStructFromBuffer(headerb, &sqdh); err!=nil {
    return nil, err
  }
  log.Printf("%#v", sqdh)
  //var body []byte
  body:=make([]byte,sqdh.MsgLength+28-266)
  f.Read(body)
  log.Printf("%s", body)
  if s.indexStructure[position-1].CRC!=bufHash32(string(sqdh.To[:])) {
    return nil, errors.New(fmt.Sprintf("Wrong message CRC need 0x%08x, got 0x%08x", s.indexStructure[position-1].CRC, bufHash32(string(sqdh.To[:]))))
  }
  rm:=&Message{}
  rm.From=strings.Trim(string(sqdh.From[:]),"\x00")
  rm.To=strings.Trim(string(sqdh.To[:]),"\x00")
  rm.FromAddr=types.AddrFromNum(sqdh.FromZone, sqdh.FromNet, sqdh.FromNode, sqdh.FromPoint)
  rm.ToAddr=types.AddrFromNum(sqdh.ToZone, sqdh.ToNet, sqdh.ToNode, sqdh.ToPoint)
  rm.Subject=strings.Trim(string(sqdh.Subject[:]),"\x00")
  rm.Body=string(body[:])
  kla:=strings.Split(rm.Body[1:sqdh.CLen],"\x01")
  rm.Body="\x01"+strings.Join(kla,"\x0d\x01")+"\x0d"+rm.Body[sqdh.CLen+1:]
  if strings.Index(rm.Body,"\x00")!=-1 {
    rm.Body=rm.Body[0:strings.Index(rm.Body,"\x00")]
  }
  err=rm.ParseRaw()
  if err!=nil {
    return nil, err
  }
  log.Printf("%#v", rm)
  return rm, nil
}

func (s *Squish) readSQI() {
  if len(s.indexStructure)>0 { return }
  file, err := os.Open(s.AreaPath+".sqi")
  if err!=nil {
    return
  }
  reader := bufio.NewReader(file)
  part := make([]byte, 12288)
  for {
    count, err := reader.Read(part);
    if err!=nil {
      break
    }
    partb:=bytes.NewBuffer(part[:count])
    for {
      var sqi sqi_s
      if err=utils.ReadStructFromBuffer(partb, &sqi); err!=nil {
        break
      }
      if sqi.Offset!=0 {
        s.indexStructure=append(s.indexStructure,sqi)
      }
    }
  }
//  log.Printf("%s %#v", s.AreaName, s.indexStructure)
}
func (s *Squish) GetLast() uint32 {
  s.readSQI()
  if len(s.indexStructure)==0 { return 0 }
  file, err := os.Open(s.AreaPath+".sql")
  defer file.Close()
  if err!=nil {
    return 0
  }
  var ret uint32
  err=binary.Read(file, binary.LittleEndian, &ret)
  if err!=nil {
    return 0
  }
  for i,is:=range s.indexStructure {
    if ret==is.MessageNum {
//      log.Printf("ret, i: %d %d",ret, i)
      return uint32(i+1)
    }
  }
  return 0
}

func (s *Squish) GetCount() uint32 {
  s.readSQI()
  return uint32(len(s.indexStructure))
}

func (s *Squish) GetType() EchoAreaType {
  return EchoAreaTypeSquish
}

func (s *Squish) Init() {
}

func (s *Squish) GetName() string {
  return s.AreaName
}

func bufHash32(str string) (h uint32) {
  h=0
  for _, b:=range strings.ToLower(str) {
    if b==0 { continue }
    h=(h<<4)+uint32(b)
    g:= h & 0xF0000000;
    if g!=0 {
      h |= g >> 24
      h |= g
    }
  }
  h=h & 0x7fffffff
  return
}
