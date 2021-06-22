package app

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/wh-servers/tcp_app/socket"
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

func RegisterHandler(handlers ...*Handler) error {
	for _, h := range handlers {
		handlerMtx.Lock()
		/*
			todo: according to different reqType, to assign to different queues,
			different queues need different worker number, important handler need more queues
		*/
		handlerMap[h.CmdNo] = h
		handlerMtx.Unlock()
	}
	return nil
}

//distinguish different cmd and use different handler
func HandlerManager(connClient *socket.ConnClient) error {
	var err error
	ticker := time.NewTicker(connClient.ConnTimeout)
	go func() {
		for {
			cmdNo, req, resp, e := unwrapMsg(connClient)
			if e != nil {
				err = e
				connClient.IsDead <- true
				break
			}
			if handler, ok := handlerMap[cmdNo]; ok {
				if e := handler.Processor(context.Background(), req, resp); e != nil {
					err = e
					connClient.IsDead <- true
					break
				}
			}
			e = wrap(resp, connClient)
			if e != nil {
				err = e
				connClient.IsDead <- true
				break
			}
			ticker.Reset(connClient.ConnTimeout)
		}
	}()
	select {
	case <-ticker.C:
		fmt.Printf("conn %v closed due to no request for %d seconds long\n", connClient, connClient.ConnTimeout/time.Second)
	case <-connClient.IsDead:
		fmt.Printf("conn %v closed due to err happen or client close\n, err: %v\n", connClient, err)
	}
	return connClient.Close()
}

func unwrapMsg(connClient *socket.ConnClient) (cmdNo uint8, req, resp interface{}, e error) {
	var msg []byte
	var err error
	if connClient == nil {
		return math.MaxUint8, nil, nil, fmt.Errorf("nil conn in app")
	}
	err = connClient.Read(&msg)
	if err != nil {
		return math.MaxUint8, nil, nil, fmt.Errorf("read msg err : %v", err)
	}
	if len(msg) < 2 {
		return math.MaxUint8, nil, nil, fmt.Errorf("wrong msg format in conn")
	}
	cmdNo = msg[0]
	var ok bool
	if handler, exist := handlerMap[cmdNo]; exist {
		//note: have to use reflect.Indirect()
		reqNewValue := reflect.New(reflect.Indirect(reflect.ValueOf(handler.Req)).Type()).Interface()
		req, ok = reqNewValue.(proto.Message)
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
		respValue := reflect.New(reflect.Indirect(reflect.ValueOf(handler.Resp)).Type()).Interface()
		resp, ok = respValue.(proto.Message)
		if !ok { //default *[]byte
			resp = respValue.(*[]byte)
		}
	} else {
		return math.MaxUint8, nil, nil, fmt.Errorf("no cmd registered")
	}
	return
}

func wrap(msg interface{}, connClient *socket.ConnClient) error {
	if connClient == nil {
		return fmt.Errorf("nil conn in app")
	}
	resp, ok := msg.(proto.Message)
	var respData []byte
	var err error
	if ok {
		respData, err = proto.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal resp err: %v", err)
		}
	} else { //default type: []byte
		if msg == nil {
			return nil
		}
		respData = *(msg.(*[]byte))
	}
	err = connClient.Write(respData)
	if err != nil {
		return fmt.Errorf("write msg to conn err: %v", err)
	}
	return err
}
