package app

import (
	"fmt"
	"os"
	"time"

	"github.com/wh-servers/tcp_app/config"
	"github.com/wh-servers/tcp_app/socket"
)

var handlerMap = map[uint8]*Handler{}

type App struct {
	Socket *socket.Socket
	isStop bool
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
	for !a.isStop {
		conn, err := a.Socket.Listener.AcceptTCP()
		if err != nil {
			fmt.Printf("accept request err: %v, connection is: %v\n", err, conn)
			continue
		}
		connClient := socket.NewConnClient()
		err = connClient.Init(a.Socket, conn)
		if err != nil {
			return err
		}
		a.Socket.ConnClientPool <- connClient
		go a.Dispatcher()
	}
	return err
}

func (a *App) Stop(s os.Signal) {
	a.isStop = true
	fmt.Println("server exit with signal: ", s)
}

//dispatch connections to different workers
//first come first out
func (a *App) Dispatcher() (err error) {
	//todo: add multi workers, now consider only one worker
	connClient := <-a.Socket.ConnClientPool
	//todo: add heartbeat
	ticker := time.Tick(connClient.ConnTimeout)
	for {
		select {
		case <-ticker:
			fmt.Printf("connection: %v closed due to timeout\n", connClient)
			connClient.IsDead <- true
			return
		case <-connClient.IsDead:
			fmt.Printf("connection: %v closed due to keep alive fail\n", connClient)
			connClient.IsDead <- true
			return
		default:
			err = HandlerManager(connClient)
		}
	}
}

func (a *App) initSocket(conf config.Config) error {
	newSocket := socket.NewSocket()
	err := newSocket.Init(conf.SocketOptions...)
	a.Socket = newSocket
	return err
}
