package app

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/golang/protobuf/proto"
)

// ProcessorFunc is registered api handler
type ProcessorFunc func(ctx context.Context, request, response interface{}) error
type Handler struct {
	CmdNo     uint8
	Processor ProcessorFunc
	Req       interface{}
	Resp      interface{}
}

var handlerMtx = sync.Mutex{}

func (a *App) RegisterHandler(handlers ...*Handler) error {
	for _, h := range handlers {
		handlerMtx.Lock()
		//reqType := reflect.TypeOf(h.Req)
		//respType := reflect.TypeOf(h.Resp)
		/*todo: according to different reqType, to assign to different queues,
		different queues need different worker number, important handler need more queues
		*/
		handlerMap[h.CmdNo] = h
		handlerMtx.Unlock()
	}
	return nil
}

//distinguish different cmd and use different handler
func (a *App) HandlerManager() error {
	cmdNo, req, resp, err := a.unwrapMsg()
	if err != nil {
		return err
	}
	//todo: use worker to process handler. worker(handlerMap[reqData[0]],reqData[1:], &respData)
	if handler, ok := handlerMap[cmdNo]; ok {
		if err := handler.Processor(context.Background(), req, resp); err != nil {
			return fmt.Errorf("process req err: ", err)
		}
	}
	err = a.wrap(cmdNo, resp)
	return err
}

func (a *App) unwrapMsg() (cmdNo uint8, req, resp interface{}, err error) {
	var msg []byte
	if a.Socket == nil || len(a.Socket.ConnClientPool) < 1 {
		return math.MaxUint8, nil, nil, fmt.Errorf("nil conn in app")
	}
	/*todo: find correct conn, now only consider one connection
	now use the last conn
	and no lock to the pool
	*/
	err = a.Socket.ConnClientPool[len(a.Socket.ConnClientPool)-1].Read(&msg)
	if err != nil {
		return math.MaxUint8, nil, nil, fmt.Errorf("read msg from conn err: ", err)
	}
	if len(msg) < 2 {
		return math.MaxUint8, nil, nil, fmt.Errorf("no msg from conn")
	}
	cmdNo = msg[0]
	//note: have to use reflect.Indirect()
	reqNewValue := reflect.New(reflect.Indirect(reflect.ValueOf(handlerMap[cmdNo].Req)).Type()).Interface()
	req, ok := reqNewValue.(proto.Message)
	if ok {
		//msg[0] is cmd number
		//msg[1:] is main msg
		err = proto.Unmarshal(msg[1:], req.(proto.Message))
		if err != nil {
			return math.MaxUint8, nil, nil, fmt.Errorf("unmarshal req error")
		}
	} else { //default type: []byte
		reqPointer := &[]byte{}
		*reqPointer = msg[1:]
		req = reqPointer
	}
	respValue := reflect.New(reflect.Indirect(reflect.ValueOf(handlerMap[cmdNo].Resp)).Type()).Interface()
	resp, ok = respValue.(proto.Message)
	if !ok { //default *[]byte
		resp = respValue.(*[]byte)
	}
	return
}

func (a *App) wrap(cmdNo uint8, msg interface{}) error {
	if a.Socket == nil || len(a.Socket.ConnClientPool) < 1 {
		return fmt.Errorf("nil conn in app")
	}
	/*todo: find correct conn, now only consider one connection
	now use the last conn
	and no lock to the pool
	*/
	resp, ok := msg.(proto.Message)
	var respData []byte
	var err error
	if ok {
		respData, err = proto.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal resp err: ", err)
		}
	} else { //default type: []byte
		respData = *(msg.(*[]byte))
	}
	err = a.Socket.ConnClientPool[len(a.Socket.ConnClientPool)-1].Write(respData)
	if err != nil {
		return fmt.Errorf("write msg to conn err: ", err)
	}
	return err
}
