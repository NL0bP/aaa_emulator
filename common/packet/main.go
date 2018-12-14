package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"net"

	"../crypt"
)

// Reader ... Reads packets
type Reader struct {
	Pack   []byte
	Offset uint16
	Len    uint16
	Err    bool
}

// CreateReader ... Creates packet reader
func CreateReader(buffer []byte) (reader *Reader) {
	reader = &Reader{Pack: buffer, Offset: 0, Len: uint16(len(buffer)), Err: false}
	return
}

// Byte .. read byte
func (pr *Reader) Byte() byte {
	if pr.Offset+1 > pr.Len {
		pr.Err = true
		return 0
	}
	defer func() {
		pr.Offset++
	}()
	return byte(pr.Pack[pr.Offset])
}

// Short ... read short
func (pr *Reader) Short() uint16 {
	if pr.Offset+2 > pr.Len {
		pr.Err = true
		return 0
	}
	defer func() {
		pr.Offset += 2
	}()
	return binary.LittleEndian.Uint16(pr.Pack[pr.Offset : pr.Offset+2])
}

// Int ... read integer
func (pr *Reader) Int() int {
	if pr.Offset+4 > pr.Len {
		pr.Err = true
		return 0
	}
	defer func() {
		pr.Offset += 4
	}()
	return int(binary.LittleEndian.Uint32(pr.Pack[pr.Offset : pr.Offset+4]))
}

// Int24 ... read integer24
func (pr *Reader) Int24() int {
	if pr.Offset+3 > pr.Len {
		pr.Err = true
		return 0
	}
	b1 := pr.Byte()
	b2 := pr.Byte()
	b3 := pr.Byte()
	var num int = int(b3)<<16 | int(b2)<<8 | int(b1)
	return num
}

// Long ... read long integer
func (pr *Reader) Long() uint64 {
	if pr.Offset+8 > pr.Len {
		pr.Err = true
		return 0
	}
	defer func() {
		pr.Offset += 8
	}()
	return binary.LittleEndian.Uint64(pr.Pack[pr.Offset : pr.Offset+8])
}

// String ... read string of provided length
func (pr *Reader) String(len uint16) string {
	if pr.Offset+len > pr.Len {
		pr.Err = true
		return ""
	}
	defer func() {
		pr.Offset += len
	}()
	return string(pr.Pack[pr.Offset : pr.Offset+len])
}

// Bytes ... return bytes of required length
func (pr *Reader) Bytes(len uint16) []byte {
	if pr.Offset+len > pr.Len {
		pr.Err = true
		return nil
	}
	defer func() {
		pr.Offset += len
	}()
	return pr.Pack[pr.Offset : pr.Offset+len]
}

// Writer ... Reads packets
type Writer struct {
	pack   *bytes.Buffer
	Offset uint16
	size   uint16
	opcode uint16
}

// CreateWriter ... PacketWriter constructor
func CreateWriter(opcode int) *Writer {
	pw := new(Writer)
	pw.pack = new(bytes.Buffer)
	binary.Write(pw.pack, binary.LittleEndian, uint16(0))
	binary.Write(pw.pack, binary.LittleEndian, uint16(opcode))
	pw.Offset = 2
	return pw
}

// Byte ... Convert byte to byter and write in pack
func (pw *Writer) Byte(data byte) {
	defer func() {
		pw.Offset++
	}()
	binary.Write(pw.pack, binary.LittleEndian, data)
}

// Short ... Convert short to byter and write in pack
func (pw *Writer) Short(data uint16) {
	defer func() {
		pw.Offset += 2
	}()
	binary.Write(pw.pack, binary.LittleEndian, data)
}

// Int ... Convert int to byter and write in pack
func (pw *Writer) Int(data int) {
	defer func() {
		pw.Offset += 4
	}()
	binary.Write(pw.pack, binary.LittleEndian, data)
}

func (pw *Writer) UInt(data uint32) {
	defer func() {
		pw.Offset += 4
	}()
	binary.Write(pw.pack, binary.LittleEndian, data)
}

func (pw *Writer) UInt24(data uint32) {
	b1 := byte(255 & data)
	b2 := byte(255 & (data >> 8))
	b3 := byte(255 & (data >> 16))
	pw.Byte(b1)
	pw.Byte(b2)
	pw.Byte(b3)
}

// Long ... Convert long to byter and write in pack
func (pw *Writer) Long(data uint64) {
	defer func() {
		pw.Offset += 8
	}()
	binary.Write(pw.pack, binary.LittleEndian, data)
}

// String ... Convert strinmain to byter and write in pack
func (pw *Writer) String(data string) {
	dataLen := uint16(len([]byte(data)))
	defer func() {
		pw.Offset += dataLen
	}()
	pw.Short(dataLen)
	binary.Write(pw.pack, binary.LittleEndian, []byte(data))
}

// Bytes ... Just writing bytes
func (pw *Writer) Bytes(data []byte) {
	dataLen := uint16(len(data))
	defer func() {
		pw.Offset += dataLen
	}()
	pw.pack.Write(data)
}

// HexString ... Write hex string into byte array
func (pw *Writer) HexString(data string) {
	strLen := uint16(len(data) / 2)
	defer func() {
		pw.Offset += strLen
	}()

	hexStr, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	binary.Write(pw.pack, binary.LittleEndian, []byte(hexStr))
}

// HexStringL ... Write hex string into byte array and add short len before
func (pw *Writer) HexStringL(data string) {
	hexStr, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	out := []byte(hexStr)
	bLen := uint16(len(out))
	pw.Short(bLen)
	defer func() {
		pw.Offset += bLen
	}()
	binary.Write(pw.pack, binary.LittleEndian, out)
}

// Send ... Send message from buffer
func (pw *Writer) Send(conn net.Conn) error {
	binary.LittleEndian.PutUint16(pw.pack.Bytes()[0:2], pw.Offset)
	_, err := conn.Write(pw.pack.Bytes())
	return err
}

// Send ... Send message from buffer
func (pw *Writer) SendRaw(conn net.Conn) error {
	_, err := conn.Write(pw.pack.Bytes()[4:])
	return err
}

// place it in Sess, because it'll work only with 1 player on server
//var encSeq = uint8(0)

// EncWriter ... Encrypted Server Packets
type EncWriter struct {
	Writer
	head   []byte
	encSeq *uint8
}

// CreateEncWriter ... EncWriter constructor
func CreateEncWriter(opcode uint16, seq *uint8) *EncWriter {
	pw := new(EncWriter)
	pw.encSeq = seq
	pw.pack = new(bytes.Buffer)
	pw.head = make([]byte, 4)
	pw.opcode = opcode
	binary.LittleEndian.PutUint16(pw.head[0:2], uint16(0))
	binary.LittleEndian.PutUint16(pw.head[2:4], uint16(1501))

	binary.Write(pw.pack, binary.LittleEndian, byte(0))        // crc8
	binary.Write(pw.pack, binary.LittleEndian, byte(0))        // seqNum
	binary.Write(pw.pack, binary.LittleEndian, uint16(opcode)) // opcode
	pw.Offset = 6
	return pw
}

// Send ... Send message from buffer encrypted
func (pw *EncWriter) Send(conn net.Conn) {
	defer func() {
		*pw.encSeq++
	}()

	pw.pack.Bytes()[1] = *pw.encSeq
	binary.LittleEndian.PutUint16(pw.pack.Bytes()[2:4], pw.opcode)
	pw.pack.Bytes()[0] = crypt.Crc8(pw.pack.Bytes()[1:])
	encData := crypt.ToClientEncr(pw.pack.Bytes())

	binary.LittleEndian.PutUint16(pw.head[0:2], pw.Offset)
	conn.Write(append(pw.head, encData...))
}

// ProxyWriter ... Encrypted Server Packets
type ProxyWriter struct {
	Writer
	head []byte
	//proxySeq *uint8
}

// CreateProxyWriter ... ProxyWriter constructor
func CreateProxyWriter(opcode uint16) *ProxyWriter {

	pw := new(ProxyWriter)
	//pw.proxySeq = seq
	pw.pack = new(bytes.Buffer)

	pw.opcode = opcode
	binary.Write(pw.pack, binary.LittleEndian, uint16(0))
	binary.Write(pw.pack, binary.LittleEndian, uint16(733))
	binary.Write(pw.pack, binary.LittleEndian, uint16(opcode))
	pw.Offset = 4
	return pw
}

// Send ... Send message from buffer as proxy type
func (pw *ProxyWriter) Send(conn net.Conn) error {
	binary.LittleEndian.PutUint16(pw.pack.Bytes()[0:2], pw.Offset)
	binary.LittleEndian.PutUint16(pw.pack.Bytes()[4:6], pw.opcode)
	_, err := conn.Write(pw.pack.Bytes())
	return err
}
