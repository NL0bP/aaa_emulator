package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"../common/packet"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	serv := Server{addr: "0.0.0.0:1250", IdleTimeout: 55000}
	err := serv.listen()
	fmt.Println("Error launching Proxy server:", err)
}

// Connection ... Class for Connection
type Connection struct {
	net.Conn
	IdleTimeout time.Duration
	buffSize    int16
}

// Server ... Class of server
type Server struct {
	addr        string
	IdleTimeout time.Duration
	//db          *gorm.DB
}

func (s Server) listen() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer func() {
		listener.Close()
	}()

	fmt.Printf("Proxy server started [%v]\n", s.addr)

	for {
		newConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}

		conn := &Connection{
			Conn:        newConn,
			IdleTimeout: s.IdleTimeout,
		}
		go handle(conn, &s)
	}
}

func handle(conn *Connection, serv *Server) {
	defer func() {
		conn.Close()
	}()
	fmt.Printf("[%v] new Connection\n", conn.RemoteAddr())

	sess := &Session{conn: conn}
	plenBuf := make([]byte, 2)

	for {
		_, err := conn.Read(plenBuf)
		if err != nil {
			fmt.Println("Packet size error:", err)
			break
			//continue
		}

		plen := binary.LittleEndian.Uint16(plenBuf)
		packBuf := make([]byte, plen)
		_, err = conn.Read(packBuf)
		if err != nil {
			fmt.Println("Packet reading error:", err)
			break
		}
		reader := packet.CreateReader(packBuf)
		opcode := reader.Short()

		switch opcode {
		case 1:
			sess.ProxyExist(reader)
		case 2:
			sess.Answer(reader)
		case 5:
			sess.Answer(reader)
		case 18:
			sess.Answer(reader)

		default:
			fmt.Println("[PROXY] No such subtype:", opcode)
		}

	}
}
