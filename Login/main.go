package main

// "runtime"
import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var loginServerInstance *LoginServer

func main() {
	if len(os.Args) > 1 {
		fmt.Print("check\n")
		result := net.ParseIP(os.Args[1])
		i, err := strconv.Atoi(os.Args[2])
		if result != nil && err != nil && i > 0 && i < 65536 {
			loginServerInstance = createLogin(os.Args[1]+":"+os.Args[2], 20)
			loginServerInstance.Listen()
			return
		}
	}

	loginServerInstance = createLogin("0.0.0.0:1237", 20)
	loginServerInstance.Listen()
}
