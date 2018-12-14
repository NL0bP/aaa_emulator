package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"../common/packet"
)

type GameServerInfo struct {
	verbose  string
	sid      byte
	stype    byte
	scolor   byte
	load     byte // 0 - low, 1 - avg, 2 - high
	isOnline byte
	ipAddr   []byte
	port     uint16
}

type LoginServer struct {
	address                string
	timeout                time.Duration
	max_pers               byte
	max_pers_expander_item byte
	gameServers            []GameServerInfo
}

func createLogin(address string, timeout int) (serv *LoginServer) {
	timeoutDur, _ := time.ParseDuration(strconv.Itoa(timeout) + "s")
	serv = &LoginServer{address: address,
		timeout:                timeoutDur,
		max_pers:               1,
		max_pers_expander_item: 1,
		gameServers:            make([]GameServerInfo, 0)}
	id := byte(0)
	//serv.gameServers = append(serv.gameServers, GameServerInfo{"SNE lab ...", id, 1, 2, 0, 1, []byte{...}, 1239})
	//id++
	serv.gameServers = append(serv.gameServers, GameServerInfo{"localhost", id, 1, 2, 0, 1, []byte{127, 0, 0, 1}, 1239})
	id++
	return
}

func (obj *LoginServer) Listen() {
	listener, _ := net.Listen("tcp", obj.address)

	defer func() { listener.Close() }()

	fmt.Printf("ArchAge login server started at %s", obj.address)
	for {
		connection, error := listener.Accept()
		if error != nil {
			fmt.Println("Error in connection establishment:\n" + error.Error())
			connection.Close()
			continue
		}

		go handleSession(connection)
	}
}

func handleSession(client net.Conn) {
	fmt.Println("Login session started for " + client.RemoteAddr().String())

	defer func() {
		client.Close()
	}()

	session := session{client: client}
	var (
		err    error
		buffer []byte
		opCode uint16
		parser *packet.Reader
	)
	//login data should be somewhere here
	for {
		err, _, buffer = readPacket(client, 0)
		if err != nil {
			fmt.Printf("[%s] %s - %s: %s\n", time.Now().String(), session.client.RemoteAddr().String(), session.username, "Error in message receiving: "+err.Error())
			return
		}

		parser = packet.CreateReader(buffer)

		opCode = parser.Short()

		switch opCode {
		case 6:
			err = session.challengeResponse2(parser)
			if err != nil {
				fmt.Printf("[%s] %s - %s: %s\n", time.Now().String(), session.client.RemoteAddr().String(), session.username, err.Error())
				return
			}
		case 0xC:
			err = session.cancelEnterWorld(parser)
			if err != nil {
				fmt.Printf("[%s] %s - %s: %s\n", time.Now().String(), session.client.RemoteAddr().String(), session.username, err.Error())
				return
			}
		case 0xD:
			err = session.requestReconnect(parser)
		}
	}
}

func readPacket(client net.Conn, timeout time.Duration) (out_err error, dataLen uint16, buffer []byte) {
	//defer readRecover(client, &exitCode)
	if timeout != 0 {
		client.SetReadDeadline(time.Now().Add(timeout))
	} else {
		client.SetReadDeadline(time.Time{})
	}
	if client == nil {
		out_err = errors.New("Client object is equal to nil")
		return
	}
	dataLenB := make([]byte, 2)
	_, err := client.Read(dataLenB)
	if err != nil {
		out_err = err
		return
	}
	dataLen = binary.LittleEndian.Uint16(dataLenB)
	if dataLen > 1 {
		buffer = make([]byte, dataLen)
		_, err = client.Read(buffer)
		if err != nil {
			out_err = err
			return
		}
	} else {
		out_err = errors.New("Wrong data size")
		return
	}
	return
}
