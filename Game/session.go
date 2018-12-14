package main

import (
	//"encoding/hex"
	"encoding/hex"
	"fmt"

	"../common/crypt"
	"../common/packet"
	//"github.com/jinzhu/gorm"
)

// Session ... Session
type Session struct {
	conn *Connection
	//db   *gorm.DB
	accountID    int
	uid          uint
	cr           *crypt.CryptAES
	kostyl       int
	alive        bool
	ingame       bool
	visibleChars []int
}

type Character struct {
	posX uint32
	posY uint32
	posZ uint32
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// X2EnterWorld ... opc=0, type(H), pFrom(I), pTo(I), accID(I), cookie(Q)
// hiiQIihIi
func (sess *Session) X2EnterWorld(reader *packet.Reader) {
	fmt.Println("Enter world")
	ttype := reader.Short()
	pFrom := reader.Int()
	pTo := reader.Int()
	accountID := reader.Int()
	cookie := reader.Long()
	zoneID := reader.Int()
	tb := reader.Short()
	revision := reader.Int()
	index := reader.Int()

	fmt.Println("[GAME, X2EnterWorld]:", ttype, pFrom, pTo, accountID, cookie,
		zoneID, tb, revision, index)

	// Check if authorized here

	// Load char data from DB

	sess.accountID = accountID

	sess.EnterWorldResponse()
	sess.ChangeState(0)
}

// Dev. name
func (sess *Session) getKeys(reader *packet.Reader, rsa crypt.CryptRSA) {
	reader.Short() // Unknown, always = 355 (0x6301)
	reader.Int()   // lenAES ?
	reader.Short() // lenXOR ?
	encAES := reader.Bytes(128)
	encXOR := reader.Bytes(128)

	aesKey := rsa.GetAesKey(encAES)
	xorKey := rsa.GetXorKey(encXOR)

	sess.cr = crypt.ClientCrypt(aesKey, xorKey)

	fmt.Println("[GAME, getKeys]: AES: ", hex.EncodeToString(aesKey), ", XOR: ", xorKey)

	// TODO: Тут эти респонсы над вызывать (pers_info функция в питоне смотри)
	/*
			# 2nd block
		    self.World_0x272()
		    self.World_0xEC(enter=True)
		    self.World_0x8c()

		    # 3rd block
		    self.World_0x14d()
			self.CharacterListPacket()
	*/

	// С нижней залупой (2nd block, 3rd block респонсы) заходит в лобби без ошибок
	// Если раскомментить BigPack, то зайдет в лобби но с ошибкой потому, что он отсылается только при входе игру
	//h, _ := hex.DecodeString("0700dd05d7bdf353102a00dd05446fdd01d3a2724213e3b3835323f4c494643405cdc5f82b16e6b6865727f7c797704010e0b08151fe00dd050df94a96663707d7a7775020f0c090613101d1a1724212e2b2835323f3c494643404d5a5754515e6b6865626f7c7976737fed0a0704011e1b1815122f2c292623303d3a3734414e4b4845525f5c596663606d6a7774717e7c0906030fed1a1714111e2b2825222f3c393633304d4a4744415e5b5855526f6c696663707d7a7774010e0b0815121f1c192623202d2a3734313e3b4845424f4c595653505d6a6764616e7b7875727fed0a0704011e1b1815122f2c292623303d3a3744414e4b4855525f5c596663606d6a7774717e7b0805020f0c191613101d2a2724212e3b3835323f4c494643405d5a5754516e6b6865727f7c797704010e0b081510f00dd050f37bac697704010e0b0815186f101dd0589d4ec6537075c3574471ae7f3ee065697b5c8142d7d8cd0ec5320f3b092633343d2a474f714e4b5e05525f5c696663606d7a7774717e0b098515ebcc1917d5102d2b9704313e3b3835425f9c994653505d1b7ac4516e6b6879f9236c4a024f140ac7a915121f1c292623202d3a3734314e4b4845425f5c595653206d6a6764717e7b7875020f0c0105e3101d1a1724292ddb283d31cf3c394643404d4a57145a9e71c865626f7c797673710e030bf502171fe9161b23d5d60477c13e333bc542474fb9465b53a36defdb9b90a5978f8cb1838c8280801d1a1714191e2b26d52cdf32d93623004d4a4744415f4b5855526f63896600cbed77f77be2024c0b87adae0d14672b225c1b2835323f3a094643404d5a5752115e6b6765626f7c7976737d505a0704075e1b18151dbf2c292823303d3a3734414e4b4845525f5c595663606d6a7774717e7c0906030fed1a1714111e2b2825222f3c393633304e0b6821ba1beb5855526f6c66039835cd7a7774010e0b081a77e459a92623202d2a373436d4077df5424f4c595653505d6a6764616e7b7875727fed0a0704011e1b1815122f2c292623303d3a3744414e4b4855525f5c596663606d6a7774717e7b0805020f0c191613101d2a2724212e3b3835323f4c494643505d5a5744516e6b6865727e65f542b4010e0b08151245484")
	//sess.conn.Write(h)

	sess.persInfo()

	//h1, _ := hex.DecodeString("2400dd0564f1fc825223f4c495643405d55a754516e6a91e947cf7c797704010e0b081514272")
	//sess.conn.Write(h1)

	//h2, _ := hex.DecodeString("1d00dd05107771045f36774517e6bd86214285b4fe1f2e30d1bd8b5dc4f4231d00dd05cd7071045f36774514e6bd86214285b4fe1f2e30d2bd8b5dc4f423")
	//sess.conn.Write(h2)
}

func (sess *Session) OnMovement(pack []byte) {
	reader := packet.CreateReader(pack)
	reader.Byte()
	reader.Byte()
	reader.Short() // op :=
	reader.Int24() // bc :=
	reader.Byte()  // _type :=
	reader.Int()   // time :=
	reader.Byte()  // flags :=

	posX := reader.Int24()
	posY := reader.Int24()
	posZ := reader.Int24()
	velX := reader.Short()
	velY := reader.Short()
	velZ := reader.Short()
	rotX := reader.Byte()
	rotY := reader.Byte()
	rotZ := reader.Byte()
	aDmX := reader.Byte()
	aDmY := reader.Byte()
	aDmZ := reader.Byte()
	reader.Byte() // aStace :=
	reader.Byte() // aAlertness :=
	reader.Byte() // aFlags :=
	fmt.Println(posX, posY, posZ)
	fmt.Println(rotX, rotY, rotZ)
	fmt.Println(aDmX, aDmY, aDmZ)
	fmt.Println(velX, velY, velZ)

	go sess.MovementReply(pack, uint32(posX), uint32(posY), uint32(posZ), uint16(rotX), uint16(rotY), uint16(rotZ))
}

func (sess *Session) MovementReply(pack []byte, x, y, z uint32, rx, ry, rz uint16) {
	for i := range sessions {
		if sessions[i].alive && sessions[i].ingame && sess != sessions[i] {
			if !intInSlice(sess.accountID, sessions[i].visibleChars) {
				sessions[i].visibleChars = append(sessions[i].visibleChars, sess.accountID)
				sessions[i].UnitState0x8d(x, y, z, rx, ry, rz, sess)
			}
			sessions[i].World_dd01_0x162(pack, sess)
		}
	}
}
