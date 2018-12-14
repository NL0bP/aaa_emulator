package main

import (
	"encoding/hex"
	"io/ioutil"
	"math"
	"strings"

	"../common/packet"
)

// EnterWorldResponse ... Reason(H) GM(B) SC(I) SP(H) WF(Q) TZ(I)
// HHI{0}s{1}sBBBBH
func (sess *Session) EnterWorldResponse() {
	w := packet.CreateEncWriter(0, sess.conn.encSeq)

	w.Short(0)    // Reason
	w.Byte(0)     // GM
	w.UInt(0)     // SC
	w.Short(1250) // SP
	w.Long(0)     // WF
	w.UInt(0)     // TZ

	n := "a38ab4ef39f6b852b8690298855a2b494d21af48b438228524db40b5abb1be65ea773f2116b65b74d113fdc3f7cf91a02e90cb858e20c6d954c46907b939ccefd8e8d8e46be96208a1aaa776a825dd2617a8fafe277032359c05aed96bb02e7d227448e81619c8e7991785a94f0330fced39c40f8dc55a001cbb9b426cbea86f"
	e := "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010001"
	w.Short(260)       // H, Public Key Size  0401 (Should be 260, else pizda)
	w.Short(128*2 + 4) //H, Pub key len (in pub key)
	w.UInt(1024)
	w.HexString(n)
	w.HexString(e)

	// IP Adress and Port of client??
	w.Byte(10)
	w.Byte(1)
	w.Byte(1)
	w.Byte(113)
	w.Short(57259)
	w.Send(sess.conn)
}

func (sess *Session) persInfo() {
	sess.World_0x272()
	sess.World_0xEC(true)
	sess.World_0x8c()
	sess.World_0x14d()
	sess.CharacterListPacket()
}

func (sess *Session) BeginGame() {
	sess.World_0x14f()
	sess.World_0x145()
}

//State Responses

func (sess *Session) World0x94() {
	w := packet.CreateEncWriter(0x94, sess.conn.encSeq)
	w.Byte(1) // send Address
	w.Byte(0) // sp Md5
	w.Byte(1) // lua Md5
	dir := "x2ui/hud"
	w.String(dir)
	w.Byte(0) // modPack
	w.Send(sess.conn)
}

func (sess *Session) World0x34() {
	//SCAccountInfoPacket
	host := "archeagegame.com"
	fset := "7f37340f79087dcb376503dea486380002e66fc7bb9b5d010001"
	w := packet.CreateEncWriter(0x34, sess.conn.encSeq)
	w.String(host)
	w.HexStringL(fset)
	w.UInt(0)    // count I
	w.UInt(0)    // initial Labor points I
	w.Byte(0)    // can place house
	w.Byte(0)    // can pay tax
	w.Byte(1)    // can use auction
	w.Byte(1)    // can trade
	w.Byte(1)    // can send mail
	w.Byte(1)    // can use bank
	w.Byte(1)    // can use copper
	w.Byte(0)    // second  password max fail count
	w.UInt(0)    // idle kick time I
	w.Byte(0)    // enable
	w.Byte(0)    // pcbang
	w.Byte(0)    // premium
	w.Byte(0)    // max characters
	w.Short(400) // honorPointDuringWarPercent
	w.Byte(0)    // ucc ver
	w.Byte(1)    // member type
	w.Send(sess.conn)
}

func (sess *Session) World0x2c3() {
	w := packet.CreateEncWriter(0x2c3, sess.conn.encSeq)
	platform_url := "https://session.draft.integration.triongames.priv"
	commerce_url := "https://archeage.draft.integration.triongames.priv/commerce/purchase/credits/purchase-credits-flow.action"
	w.Byte(1) // Activate
	w.String(platform_url)
	w.String(commerce_url)
	w.Short(0) // HaveWikiUrl
	w.Short(0) // HaveCsUrl
	w.Send(sess.conn)
}

func (sess *Session) World_0xEC(enter bool) {
	//SCUpdatePremiumPointPacket
	w := packet.CreateEncWriter(0xEC, sess.conn.encSeq)
	var (
		v1 uint32
		v2 uint32
		v3 uint64
		v4 uint32
		v5 uint32
	)
	if enter != true {
		v1 = 1
		v2 = 1
		v3 = 0x6E8D6018
		v4 = 0
		v5 = 0
	} else {
		v1 = 0
		v2 = 0
		v3 = 21600
		v4 = 53
		v5 = 51
	}
	w.UInt(v1) //payMethod
	w.UInt(v2) //payLocation
	w.Long(0)  //payStart
	w.Long(v3) //payEnd
	w.UInt(v4) //realPayTime ?
	w.UInt(v5) //realPayTime ?
	w.UInt(0)  //buyPremiumCount
	w.Send(sess.conn)
}

func (sess *Session) World_0x281() {
	//SCChatSpamDelayPacket
	w := packet.CreateEncWriter(0x281, sess.conn.encSeq)
	applyConfig := "0f"
	detectConfig := "000070420500000000001644cdcc4c3f0ac803"

	w.Byte(2)   //version
	w.Short(60) //report delay
	// chatTypeGroup (loop)
	w.HexString("0101010100000100000000010000010000")
	// chatGroupDelay (loop)
	w.HexString("0000000000004040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	w.Byte(0) // whisperChatGroup
	w.HexStringL(applyConfig)
	w.HexStringL(detectConfig)
	w.Send(sess.conn)
}

func (sess *Session) World_0xBA() {

	w := packet.CreateEncWriter(0xBA, sess.conn.encSeq)
	w.Byte(0) //
	w.Byte(1) //
	w.Byte(0) //
	w.Send(sess.conn)
}

func (sess *Session) World_0x18a() {
	//??
	w := packet.CreateEncWriter(0x18a, sess.conn.encSeq)
	w.Byte(0)                                         // searchLevel
	w.Byte(10)                                        // bidLevel
	w.Byte(0)                                         // postLevel
	w.Byte(0)                                         // trade
	w.Byte(0)                                         // mail
	w.HexString("000f0f0f00000f000000000000000f0000") // limitLevels (loop)
	w.Send(sess.conn)
}

func (sess *Session) World_0x1cc() {
	//convertRatioToAAPoint
	w := packet.CreateEncWriter(0x1cc, sess.conn.encSeq)
	w.Long(0)
	w.Send(sess.conn)
}

func (sess *Session) World_0x30() {
	//?
	w := packet.CreateEncWriter(0x30, sess.conn.encSeq)
	w.Byte(1) // ingameShopVersion
	w.Byte(2) // secondPriceType
	w.Byte(0) // askBuyLaborPowerPotion
	w.Send(sess.conn)
}

func (sess *Session) World_0x1af() {
	// SCConflictZoneStatePacket
	w := packet.CreateEncWriter(0x1af, sess.conn.encSeq)
	w.UInt(0) //indunCount
	//{type 2 pvp 1 duel 1}
	w.UInt(0) //conflictCount
	//{type 2 peaceMin 4}
	w.Send(sess.conn)
}

// World_0x2cf ... World date
func (sess *Session) World_0x2cf() {
	w := packet.CreateEncWriter(0x2cf, sess.conn.encSeq)
	w.Byte(1) //protectFaction
	w.Long(0) //time
	w.UInt(0) //Year
	w.UInt(0) //Month
	w.UInt(0) //Day
	w.UInt(0) //Hour
	w.UInt(0) //Min
	w.Send(sess.conn)
}

func (sess *Session) World_0x29c() {
	//SCDominionDataPacket
	w := packet.CreateEncWriter(0x1cc, sess.conn.encSeq)
	w.UInt(0) // {type declareDominion}
	w.Send(sess.conn)
}

func (sess *Session) World_6_BigPacket() {
	var (
		data []byte
		err  error
	)
	if sess.accountID == 1 {
		data, err = ioutil.ReadFile("etc/big_bad")
	} else {
		data, err = ioutil.ReadFile("etc/big_bad3")
	}
	if err != nil {
		panic(err)
	}
	sess.conn.Write(data)
}

func (sess *Session) World_0x272() {
	w := packet.CreateEncWriter(0x272, sess.conn.encSeq)
	w.Byte(0)
	w.Send(sess.conn)
}

func (sess *Session) World_0x8c() {
	w := packet.CreateEncWriter(0x8c, sess.conn.encSeq)
	data := strings.Repeat("00", 248)
	w.HexString(data)
	w.Send(sess.conn)
}

func (sess *Session) World_0x14d() {
	w := packet.CreateEncWriter(0x14d, sess.conn.encSeq)
	w.Long(0)
	w.Byte(0)
	w.Send(sess.conn)
}

func (sess *Session) CharacterListPacket() {
	w := packet.CreateEncWriter(0x79, sess.conn.encSeq)

	var charName string
	if sess.accountID == 2 {
		charName = "Rivestshamiradlemn"
	} else {
		charName = "Diffiehellman"
	}
	msg := "Hello"
	w.Byte(1)          //LastChar
	w.Byte(1)          //TotalCount
	w.UInt(0x2938)     //CharID 2938
	w.String(charName) //CharName
	w.Byte(1)          //Race
	w.Byte(2)          //Gender
	w.Byte(1)          //Level
	w.UInt(370)        //HP
	w.UInt(320)        //MP
	w.UInt(179)        //zone_id
	w.UInt(101)        //F(r)actionId
	w.String(msg)      //msg
	w.UInt(0)          //type
	w.UInt(0)          //family
	w.UInt(0x1180000)  //validFlags

	//Appearance
	w.UInt(0x4d7f)                     //
	w.UInt(0x631c)                     //
	w.UInt(0x21b)                      // HairColor
	w.UInt(0)                          // twoToneHair
	w.UInt(0xd0d01)                    // twoToneFirstWidth
	w.UInt(0x4000000)                  // twoToneSecondWidth
	w.UInt(0x3da12)                    //
	w.UInt(0xc8000000)                 //
	w.UInt(0x3c1b5)                    //
	w.UInt(math.Float32bits(0x342c54)) // Float???????????

	w.HexString("cb10000000000000000000000000000000000000000000000400000000000000000000000000803f000000000000803f0000803f00000000000000000400bc01aa00000000000000000000803f0000803f0000803f8fc2353f0000803f0000803f0000803fe37b8bffafecefffafecefff584838ff00000000800000ef00ef00ee000103000000000000110000000000fe00063bb900d800ee00d400281bebe100e700f037230000000000640000000000000064000000f0000000000000002bd50000006400000000f9000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	//Pers Info
	w.Short(0x1234)    //LaborPower
	w.Long(0x5bb45ff6) //lastLaborPowerModified
	w.Short(0)         //DeathCount
	w.Long(0x5bb45ff6) //deadTime
	w.UInt(0)          //rezWaitDuration
	w.Long(0x5bb45ff6) //rezTime
	w.UInt(0)          //rezPenaltyDuration
	w.Long(0x5bb45ff6) //lastWorldLeaveTime
	w.Long(0)          //moneyAmount
	w.Long(0)          //moneyAmount
	w.Short(0)         //crimePoint
	w.UInt(0)          //crimeRecord
	w.Short(0)         //crimeScore
	w.Long(0)          //deleteRequestedTime
	w.Long(0)          //transferRequestedTime
	w.Long(0)          //deleteDelay
	w.UInt(0)          //consumedLp
	w.Long(0)          //bmPoint
	w.Long(0)          //moneyAmount
	w.Long(0)          //moneyAmount
	w.Byte(0)          //autoUseAApoint
	w.UInt(1)          //prevPoint
	w.UInt(1)          //point
	w.UInt(0)          //gift
	w.Long(0x5bc39811) // updated
	w.Byte(0)          //forceNameChange
	w.UInt(0)          //highAbilityRsc

	w.Send(sess.conn)
}

func (sess *Session) World_0x14f() {
	w := packet.CreateEncWriter(0x14f, sess.conn.encSeq)
	w.UInt(1)
	w.Byte(1)
	w.UInt(0)
	w.Byte(255)
	w.UInt(0)
	w.Long(0)
	w.Long(0)
	w.Send(sess.conn)
}

func (sess *Session) World_0x145() {
	w := packet.CreateEncWriter(0x145, sess.conn.encSeq)
	w.UInt(0x2938) // charID
	w.Short(1)
	w.String("version 1\r\n")
	w.UInt(0xc)
	w.Send(sess.conn)
}

//Movement Packet
func (sess *Session) World_dd01_0x162(pack []byte, senderSess *Session) {
	print(hex.EncodeToString(pack[4:37]))
	w := packet.CreateWriter(0x1dd)
	w.Short(0)
	w.Short(0x162)
	w.Short(1)
	w.UInt24(0x66db + uint32(senderSess.accountID))
	w.Bytes(pack[7:38])

	w.Send(sess.conn)
}

func (sess *Session) MovePlayer(bc, x, y, z uint32) {
	w := packet.CreateWriter(0x1dd)
	w.Short(0)
	w.Short(0x162)
	w.Short(1)
	w.UInt24(bc)
	w.Byte(1)   // type
	w.UInt(0)   //tine?
	w.Byte(0)   //flags
	w.UInt24(x) //pos
	w.UInt24(y)
	w.UInt24(z)
	w.Short(0) //vel xyz
	w.Short(0)
	w.Short(0)
	w.Byte(0) //rot
	w.Byte(0)
	w.Byte(0)
	w.UInt24(0) // a.dm.xyz
	w.Byte(2)   //a.stace
	w.UInt24(0) //
	//w.Bytes(pack[7:38])

	w.Send(sess.conn)
}

//?        id     type time     fg pos XYZ              vel XYZ         rot XYZ   a.dmxyz   a.stace,alertness,flags   ???
//7d148400 a52b01 01   3f400800 00 011d7b c4db77 ae0703 0000 0000 1efd  00 00 39  00 00 00  02 00 00

//Display Unit
func (sess *Session) UnitState0x8d(x, y, z uint32, rx, ry, rz uint16, senderSess *Session) {
	w := packet.CreateEncWriter(0x8d, sess.conn.encSeq)
	w.UInt24(0x66db + uint32(senderSess.accountID)) // LiveID

	if senderSess.accountID == 1 {
		w.String("Diffiehellman") // name
	} else {
		w.String("RivestShamirAdlemn") // name
	}
	w.Byte(0) // type 0 - player
	if senderSess.accountID == 1 {
		w.UInt(0x2938) //charID
	} else {
		w.UInt(0x2b086) // charID
	}
	w.Long(0)   //something... "V"
	w.Short(0)  // String "master"
	w.UInt24(x) // coords
	w.UInt24(y)
	w.UInt24(z)
	w.UInt(0x3f800000) //Scale
	w.Byte(1)          //Level
	/*
		w.UInt(0x0B000000) // ModelRef

		//Inventory
		w.HexString("62450000000000000000000000000000005363000000000000000000000000000000E0600000000000000000000000514900000000000000000000004863000000000000000000000000000000000000000000000000000000000000000000000000000000000000002607000000000000000000000000000000D9360000000000000000000000000000007E4D0000425E00000000000000000000000000001802000000000000000000000000000003AA0E00000100000000000000000000000000803F0000803F0000000000000000000000000000803F000000000000803F350200000000803F000000000000803F0000000021000000000000003CDA3C3FFFCDC2FFA25F42FFA25F42FF2B250DFF4B4756FF800000FAFDE6F7DFE4553AF82622176437F5009CD934D8FE090800EBF06220BA2325F30E14FDFF02F0DA0FF325D7F516EB0A25C141E1B0D3159CCE0F0315001EFEF545E601043C1427FFED430DD5272A140023FCCB000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		w.UInt(10)      //HP
		w.UInt(10)      //MP
		w.Short(0xffff) //Points?
		w.Byte(0)       //isLooted
		w.Byte(0)       //activeWeapon
		w.Byte(0)       //learned SkillCount
		//w.UInt(0x28AB)  //type
		//w.Byte(1)       // level
		//w.UInt(0x2A00)  // type
		//w.Byte(1)       // level
		w.UInt(0)   // learnedBuffs
		w.Short(rx) //rotation xyz
		w.Short(ry)
		w.Short(rz)
		w.HexString("0800A662" +
			"01000000")
		//factionID confirm
		//ns.Write(npc.FactionId);
		w.UInt(0x65)

		w.HexString("0000000000000000")
	*/

	//	if senderSess.accountID == 1 {
	//		w.HexString(strings.Replace("00ffffffff0a000000000018017e4d0000455e0000180200000000000003dd02000000000000000000000000000000000000000000000100000000000000000000000000803f000000000000803f0000803f00000000000000005000003002aa0200000000001d000000803f0000803f0000803f0000803f0000803f0000803f0000803f000000005ab5f8ff5ab5f8ff3c2300ff603e48ff800000f5000011dc000b00000000170000000000f323000000003d0000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000088900000007d0000ffff00000001000000000001d4460000eb110000000000006500000000000000000030000000000000000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff01010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e3c32002864070001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010100000000db66000000000001010008007264310000020b127a01000000011500000000db660086b00200010100040086410000012539010000", "86b00200", "38290000", 1))
	//	} else {
	w.HexString("00ffffffff0a000000000018017e4d0000455e0000180200000000000003dd02000000000000000000000000000000000000000000000100000000000000000000000000803f000000000000803f0000803f00000000000000005000003002aa0200000000001d000000803f0000803f0000803f0000803f0000803f0000803f0000803f000000005ab5f8ff5ab5f8ff3c2300ff603e48ff800000f5000011dc000b00000000170000000000f323000000003d0000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000088900000007d0000ffff00000001000000000001d4460000eb110000000000006500000000000000000030000000000000000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff00000000ff01010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e3c32002864070001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010100000000db66000000000001010008007264310000020b127a01000000011500000000db660086b00200010100040086410000012539010000")
	//	}

	w.Send(sess.conn)
}
