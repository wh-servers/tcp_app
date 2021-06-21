package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wh-servers/tcp_app/app"
	"github.com/wh-servers/tcp_app/config"
	app_pb "github.com/wh-servers/tcp_app/gen"
)

var (
	conf       = flag.String("conf", "config.yml", "config")
	addr       = flag.String("addr", ":8889", "listen addr")
	sigChannel = make(chan os.Signal, 1)
)

func main() {
	flag.Parse()
	newApp := app.NewApp()
	configure := config.NewConfig()
	err := configure.Load(*conf)
	fmt.Println("loaded config, err: ", err)
	err = newApp.Init(configure)
	fmt.Println("inited app, err: ", err)
	err = app.RegisterHandler(
		&app.Handler{
			CmdNo:     uint8(app_pb.CmdNo_mock_0),
			Processor: Mock_0_Process,
			Req:       &[]byte{},
			Resp:      &[]byte{},
		},
		&app.Handler{
			CmdNo:     uint8(app_pb.CmdNo_mock_1),
			Processor: Mock_1_Process,
			Req:       &[]byte{},
			Resp:      &[]byte{},
		},
		&app.Handler{
			CmdNo:     uint8(app_pb.CmdNo_mock_2),
			Processor: Mock_2_Process,
			Req:       &app_pb.Mock2Request{},
			Resp:      &app_pb.Mock2Response{},
		},
	)
	fmt.Println("registered hanlder, err: ", err)
	go newApp.Run(*addr)
	signal.Notify(sigChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	s := <-sigChannel
	// exit program
	newApp.Stop(s)
	fmt.Println(err)
}
