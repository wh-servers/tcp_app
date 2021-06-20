package main

import (
	"flag"
	"fmt"

	"github.com/golang/protobuf/proto"
	app_pb "github.com/wh-servers/tcp_app/gen"
	"github.com/wh-servers/tcp_app/socket"
)

var (
	addr = flag.String("addr", ":8889", "listen addr")
	cmd  = flag.Uint("cmd", 2, "cmd number")
)

type requestHandler func(cmdNo uint8, req *[]byte) error

var requestHandlerMap map[uint8]requestHandler

func main() {
	flag.Parse()
	//prepare socket
	skt := socket.NewSocket()
	defer skt.Close()
	readTimeoutOption := socket.ReadTimeoutOption{ReadTimeout: 10}
	writeTimeoutOption := socket.WriteTimeoutOption{WriteTimeout: 3}
	err := skt.Init(&readTimeoutOption, &writeTimeoutOption)
	fmt.Println("inited socket, err: ", err)
	err = skt.Dial(*addr)
	fmt.Println("socket dial target, err: ", err)
	//register cmd handlers
	err = registerRequestHandler()
	fmt.Println("registerRequestHandler err: ", err)
	//call server
	var req []byte
	if handler, ok := requestHandlerMap[uint8(*cmd)]; ok {
		err = handler(uint8(*cmd), &req)
		fmt.Println("req handler err: ", err)
	}
	err = skt.ConnClientSingle.Write(req)
	fmt.Println("wrote req, err: ", err)
	//receive from server
	var feedback []byte
	err = skt.ConnClientSingle.Read(&feedback)
	fmt.Printf("read res: %s, err: %v", string(feedback), err)
}

func registerRequestHandler() error {
	requestHandlerMap = make(map[uint8]requestHandler, 0)
	requestHandlerMap[0] = req_0_handler
	requestHandlerMap[1] = req_1_handler
	requestHandlerMap[2] = req_2_handler
	return nil
}

func req_0_handler(cmdNo uint8, req *[]byte) error {
	msg := "a msg from client 0"
	msgByte := []byte(msg)
	*req = append(*req, byte(cmdNo))
	*req = append(*req, msgByte...)
	return nil
}

func req_1_handler(cmdNo uint8, req *[]byte) error {
	msg := "a msg from client 1"
	msgByte := []byte(msg)
	*req = append(*req, byte(cmdNo))
	*req = append(*req, msgByte...)
	return nil
}

func req_2_handler(cmdNo uint8, req *[]byte) error {
	reqProto := &app_pb.Mock2Request{
		Id:      333,
		Keyword: "mock 2 request",
	}
	reqData, err := proto.Marshal(reqProto)
	if err != nil {
		return err
	}
	*req = append(*req, byte(cmdNo))
	*req = append(*req, reqData...)
	return err
}
