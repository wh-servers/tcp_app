package main

import (
	"flag"
	"fmt"

	"github.com/wh-servers/tcp_app/app"
	"github.com/wh-servers/tcp_app/config"
)

var (
	conf = flag.String("conf", "config.yml", "config")
	addr = flag.String("addr", ":8889", "listen addr")
)

func main() {
	flag.Parse()
	newApp := app.NewApp()
	defer newApp.Stop()
	configure := config.NewConfig()
	err := configure.Load(*conf)
	fmt.Println("loaded config, err: ", err)
	err = newApp.Init(configure)
	fmt.Println("inited app, err: ", err)
	err = newApp.RegisterHandler(
		&app.Handler{
			//todo: cmdNo use proto enum
			CmdNo:     uint8(1),
			Processor: Mock_1_Process,
			Req:       []byte{},
			Resp:      []byte{},
		},
		&app.Handler{
			CmdNo:     uint8(2),
			Processor: Mock_2_Process,
			Req:       []byte{},
			Resp:      []byte{},
		},
	)
	fmt.Println("registered hanlder, err: ", err)
	err = newApp.Run(*addr)
	fmt.Println(err)
}
