package main

import (
	"flag"
	"fmt"

	"github.com/wh-servers/tcp_app/socket"
)

var (
	addr = flag.String("addr", ":8889", "listen addr")
)

func main() {
	flag.Parse()
	skt := socket.NewSocket()
	defer skt.Close()
	readTimeoutOption := socket.ReadTimeoutOption{ReadTimeout: 10}
	writeTimeoutOption := socket.WriteTimeoutOption{WriteTimeout: 3}
	err := skt.Init(&readTimeoutOption, &writeTimeoutOption)
	fmt.Println("inited socket, err: ", err)
	err = skt.Dial(*addr)
	fmt.Println("socket dial target, err: ", err)
	msg := "a msg from client"
	msgByte := []byte(msg)
	cmdNo := uint8(1)
	var req []byte
	req = append(req, byte(cmdNo))
	req = append(req, msgByte...)
	err = skt.ConnClientSingle.Write(req)
	fmt.Println("wrote req, err: ", err)
	var feedback []byte
	err = skt.ConnClientSingle.Read(&feedback)
	fmt.Printf("read res: %s, err: %v", string(feedback), err)
}
