package app

import (
	"context"
	"fmt"
)

// ProcessorFunc is registered api handler
type ProcessorFunc func(ctx context.Context, request, response interface{}) error
type Handler struct {
	CmdNo     uint8
	Processor ProcessorFunc
	Req       interface{}
	Resp      interface{}
}

func (a *App) RegisterHandler(handlers ...Handler) error {
	for _, h := range handlers {
		a.HandlerMap[h.CmdNo] = h.Processor
	}
	return nil
}

//distinguish different cmd and use different handler
func (a *App) HandlerManager() error {
	if a.Socket == nil || len(a.Socket.ConnClientPool) < 1 {
		return fmt.Errorf("nil conn in app")
	}
	/*todo: find correct conn, now only consider one connection
	now use the last conn
	and no lock to the pool
	*/
	conn := a.Socket.ConnClientPool[len(a.Socket.ConnClientPool)-1]
	var reqData []byte
	var respData []byte
	err := conn.Read(&reqData)
	if err != nil {
		return err
	}
	if len(reqData) < 2 {
		return fmt.Errorf("no msg in conn")
	}
	//reqData[0] is cmd number
	//reqData[1:] is main msg
	err = a.HandlerMap[reqData[0]](context.Background(), reqData[1:], &respData)
	if err != nil {
		return fmt.Errorf("process req err: ", err)
	}
	err = conn.Write(respData)
	if err != nil {
		return fmt.Errorf("write res to conn err: ", err)
	}
	return err
}
