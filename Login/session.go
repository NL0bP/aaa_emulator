package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"

	"../common/packet"
)

// Session ... Session
type session struct {
	client   net.Conn
	username string
	acc_id   uint
}

//Requests

func (sess *session) challengeResponse2(parser *packet.Reader) error {
	int1 := parser.Int()
	int2 := parser.Int()
	byte1 := parser.Byte()
	str1len := parser.Short()
	str1 := parser.String(str1len)
	loginLen := parser.Short()
	login := parser.String(loginLen)
	tokenLen := parser.Short()
	token := parser.Bytes(tokenLen)

	if parser.Err {
		return errors.New("Parse error")
	}

	fmt.Printf("Check. int1 %d, int2 %d, byte1 %d, str1len %d, str1 %s, login %s, token %s\n", int1, int2, byte1, str1len, str1, login, hex.EncodeToString(token))

	//should be proper check for login
	if login != "admin" && login != "user" {
		err := sess.loginDenied("User doesn't exists", 2)
		if err != nil {
			return err
		}
		return nil
	}

	sess.username = login

	//should be proper check for token/password
	if hex.EncodeToString(token) == "0102030405060708090a0b0c0d0e0f1000000000000000000000000000000000" {
		err := sess.loginDenied("Wrong password", 4)
		if err != nil {
			return err
		}
		return nil
	}

	err := sess.joinResponse()
	if err != nil {
		return err
	}
	err = sess.authResponse()
	if err != nil {
		return err
	}
	return nil
}

func (sess *session) cancelEnterWorld(parser *packet.Reader) error {
	err := sess.worldListPacket()
	if err != nil {
		return err
	}
	return nil
}

func (sess *session) requestReconnect(parser *packet.Reader) error {
	parser.Long()
	serverID := parser.Byte()
	sess.worldCookiePacket(rand.Uint32(), &loginServerInstance.gameServers[serverID])
	return nil
}

// Responses

func (sess *session) loginDenied(responseVerbose string, reason byte) error {
	serial := packet.CreateWriter(12)
	serial.Byte(reason)
	serial.Short(0)
	serial.String(responseVerbose)
	serial.Send(sess.client)
	err := serial.Send(sess.client)
	return err
}

func (sess *session) joinResponse() error {
	serial := packet.CreateWriter(0)
	serial.Byte(1)       // AuthID
	serial.Short(0)      // Reason
	serial.Long(4719366) // "afs" from archerage
	err := serial.Send(sess.client)
	return err
}

func (sess *session) authResponse() error {
	serial := packet.CreateWriter(3)
	if sess.username == "user1" {
		serial.Long(1) // AccountID
	} else {
		serial.Long(2)
	}
	serial.String("FE4E6C87FB6C1625CA3832B478E2E2F0")
	serial.Byte(5)
	err := serial.Send(sess.client)
	if err != nil {
		return err
	}
	return nil
}

func (sess *session) worldListPacket() error {
	serial := packet.CreateWriter(8)
	serial.Byte(byte(len(loginServerInstance.gameServers)))
	for i := range loginServerInstance.gameServers {
		serial.Byte(loginServerInstance.gameServers[i].sid)
		serial.Byte(loginServerInstance.gameServers[i].stype)
		serial.Byte(loginServerInstance.gameServers[i].scolor)
		serial.String(loginServerInstance.gameServers[i].verbose)
		serial.Byte(loginServerInstance.gameServers[i].isOnline)
		serial.Byte(loginServerInstance.gameServers[i].load)
		serial.Byte(3) // ?
		serial.Byte(0) // Humans
		serial.Byte(3) // ?
		serial.Byte(0) // Dwarfs
		serial.Byte(0) // Elfs
		serial.Byte(0) // Hari...
		serial.Byte(0) // Cats
		serial.Byte(3) // ?
		serial.Byte(0) // Warlocks
	}
	serial.Byte(0) // Char Count
	err := serial.Send(sess.client)
	//should be characters info

	return err
}

func (sess *session) worldCookiePacket(cookie uint32, gameServer *GameServerInfo) error {
	serial := packet.CreateWriter(0xA)
	serial.UInt(cookie)
	serial.Byte(gameServer.ipAddr[3])
	serial.Byte(gameServer.ipAddr[2])
	serial.Byte(gameServer.ipAddr[1])
	serial.Byte(gameServer.ipAddr[0])
	serial.Short(gameServer.port)
	serial.Long(0)
	serial.Long(0)
	serial.Short(0)

	err := serial.Send(sess.client)
	return err
}
