package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"../common/crypt"
	"../common/packet"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	serv := Server{addr: "0.0.0.0:1239", IdleTimeout: 55000}
	err := serv.listen()
	fmt.Println("Error launching Game server:", err)
}

// Connection ... Class for Connection
type Connection struct {
	net.Conn
	IdleTimeout time.Duration
	buffSize    int16
	encSeq      *uint8
	//proxySeq    *uint8
}

var sessions []*Session

// Server ... Class of server
type Server struct {
	addr        string
	IdleTimeout time.Duration
	rsa         *crypt.CryptRSA
	//db          *gorm.DB
}

func (s Server) listen() error {
	/*
		db, err := gorm.Open("sqlite3", "test.db")
		db.AutoMigrate(&Account{}, &Group{}, &GroupMemebers{})
		s.db = db
		if err != nil {
			fmt.Println("Failed connect to database:", err)
			return err
		}
	*/
	s.rsa = crypt.LoadRSA()

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer func() {
		listener.Close()
		//db.Close()
	}()

	fmt.Printf("Game server started [%v]\n", s.addr)

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
		var num uint8 = 0
		conn.encSeq = &num
		go handle(conn, &s)
	}
}

func handle(conn *Connection, serv *Server) {
	fmt.Printf("[%v] new Connection\n", conn.RemoteAddr())

	sess := &Session{conn: conn, kostyl: 1, alive: true, ingame: false}

	defer func() {
		conn.Close()
		sess.alive = false
		sess.ingame = false
	}()

	sessions = append(sessions, sess)

	plenBuf := make([]byte, 2)

	for {
		_, err := conn.Read(plenBuf)
		if err != nil {
			fmt.Println("Packet size error:", err)
			break
		}

		plen := binary.LittleEndian.Uint16(plenBuf)
		packBuf := make([]byte, plen)
		_, err = conn.Read(packBuf)
		if err != nil {
			fmt.Println("Packet reading error:", err)
			break
		}

		reader := packet.CreateReader(packBuf)
		reader.Byte()
		subtype := reader.Byte()

		switch subtype {
		case 1:
			opcode := reader.Short()
			switch opcode {
			case 0:
				sess.X2EnterWorld(reader)
			case 0xe17b:
				sess.getKeys(reader, *serv.rsa)
				//fmt.Println("0xe17b: getKeys, pers_info", opcode)
			default:
				fmt.Println("[WORLD] No opcode found:", opcode)
			}

		case 2:
			opcode := reader.Short()
			switch opcode {
			case 1:
				sess.FinishState(reader)
			case 18:
				sess.Pong(reader)
			default:
				fmt.Println("[PROXY] No opcode found:", opcode)
			}

		case 3:
			opcode := reader.Short()
			//reader := &packet.PacketReader{Pack: packBuf[4 : plen+4], Offset: 0}
			switch opcode {
			default:
				fmt.Println("[COMPRSSED] No opcode found:", opcode)
			}

		case 4:
			opcode := reader.Short()
			//reader := &packet.PacketReader{Pack: packBuf[4 : plen+4], Offset: 0}
			switch opcode {
			default:
				fmt.Println("[COMPR-MULTI] No opcode found:", opcode)
			}

		case 5:
			//reader := &packet.PacketReader{Pack: packBuf[4 : plen+4], Offset: 0}
			decr := sess.cr.Decrypt(packBuf[2:], len(packBuf))
			//seq := decr[0]  // seq?
			//hash := decr[1] // hash?
			opcode := binary.LittleEndian.Uint16(decr[2:4])
			//fmt.Printf("[%v] %v\n", sess.kostyl, hex.EncodeToString(decr))

			switch opcode {
			case 0x84:
				sess.OnMovement(decr)
			default:
				//sess.World_6_BigPacket()
				switch sess.kostyl {
				case 1:
					//sess.BeginGame()
					data, _ := hex.DecodeString("2400dd0564f1fc825223f4c495643405d55a754516e6a91e947cf7c797704010e0b081514272")
					sess.conn.Write(data)
					data, _ = hex.DecodeString("1d00dd05107771045f36774517e6bd86214285b4fe1f2e30d1bd8b5dc4f4231d00dd05cd7071045f36774514e6bd86214285b4fe1f2e30d2bd8b5dc4f423")
					sess.conn.Write(data)
					print("1\n")
				case 3:
					data, _ := hex.DecodeString("0c00dd05f26537116a238351c6f7")
					sess.conn.Write(data)
					print("3\n")
				case 4:
					data, _ := hex.DecodeString("1000dd05f631c9c797704010e0b0815186b7")
					sess.conn.Write(data)
					print("4\n")
				case 6:
					sess.World_6_BigPacket()
					print("6\n")
				default:
					fmt.Printf("[%v] %v\n", sess.kostyl, hex.EncodeToString(decr))
				}
				//fmt.Println("[WORLD-ENCR] No opcode found:", opcode)
			}
			sess.kostyl++
		default:
			fmt.Println("[GAME] No such subtype:", subtype)
		}
	}
}
