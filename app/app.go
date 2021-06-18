package app

import (
	"fmt"

	"github.com/wh-servers/tcp_app/config"
	"github.com/wh-servers/tcp_app/socket"
)

var handlerMap = map[uint8]*Handler{}

type App struct {
	Socket *socket.Socket
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init(conf config.Config) error {
	err := a.initSocket(conf)
	if err != nil {
		return err
	}
	return err
}

func (a *App) Run(addr string) error {
	if a.Socket == nil {
		return fmt.Errorf("no socket inited")
	}
	err := a.Socket.Listen(addr)
	for {
		conn, err := a.Socket.Listener.Accept()
		if err != nil {
			fmt.Printf("accept request err: %v, connection is: %v\n", err, conn)
			continue
		}
		newConnClient := socket.ConnClient{
			Conn:         conn,
			ReadTimeout:  a.Socket.ReadTimeout,
			WriteTimeout: a.Socket.WriteTimeout,
		}
		//todo: reuse conn position
		a.Socket.ConnClientPool = append(a.Socket.ConnClientPool, newConnClient)
		go a.Dispatcher()
	}
	return err
}

func (a *App) Stop() {
	//todo, stop gracefully
}

//dispatch connections to different workers
//first come first out
func (a *App) Dispatcher() error {
	//todo: add multi workers, now consider only one worker
	err := a.HandlerManager()
	return err
}

func (a *App) initSocket(conf config.Config) error {
	newSocket := socket.NewSocket()
	err := newSocket.Init(conf.SocketOptions...)
	a.Socket = newSocket
	return err
}
