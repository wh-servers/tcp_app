package app

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"

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
func HandlerManager(connClient *socket.ConnClient) error {
	cmdNo, req, resp, err := unwrapMsg(connClient)
	if err != nil {
		return err
	}
	//todo: use worker to process handler. worker(handler,req, resp)
	if handler, ok := handlerMap[cmdNo]; ok {
		if err := handler.Processor(context.Background(), req, resp); err != nil {
			return fmt.Errorf("process req err: %v", err)
		}
	}
	err = wrap(resp, connClient)
	return err
}

func unwrapMsg(connClient *socket.ConnClient) (cmdNo uint8, req, resp interface{}, err error) {
	var msg []byte
	if connClient == nil {
		return math.MaxUint8, nil, nil, fmt.Errorf("nil conn in app")
	}
	err = connClient.Read(&msg)
	if err != nil {
		return math.MaxUint8, nil, nil, fmt.Errorf("read msg from conn err: %v", err)
	}
	if len(msg) < 2 {
		return math.MaxUint8, nil, nil, fmt.Errorf("no msg from conn")
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
		respData = *(msg.(*[]byte))
	}
	err = connClient.Write(respData)
	if err != nil {
		return fmt.Errorf("write msg to conn err: %v", err)
	}
	return err
}
