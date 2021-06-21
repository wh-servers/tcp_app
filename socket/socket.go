package socket

import (
	"net"
	"time"
)

var defaultPoolSize int32 = 16
var defaultReadTimeout = 10 * time.Second
var defaultConnTimeout = 60 * time.Second

type Socket struct {
	//ConnClientPool usage: can use load balancer to keep connection live
	ConnClientPool chan *ConnClient
	PoolSize       int32
	Listener       net.Listener
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ConnTimeout    time.Duration
}

func NewSocket() *Socket {
	return &Socket{}
}

func (s *Socket) Init(options ...Option) error {
	for _, opt := range options {
		if err := opt.Apply(s); err != nil {
			return err
		}
	}
	//set pool size
	if s.PoolSize <= 0 || s.PoolSize > defaultPoolSize {
		s.PoolSize = defaultPoolSize
	}
	s.ConnClientPool = make(chan *ConnClient, s.PoolSize)
	//set conn timeout
	if s.ReadTimeout <= 0 || s.ReadTimeout > defaultReadTimeout {
		s.ReadTimeout = defaultReadTimeout
	}
	if s.ConnTimeout <= 0 || s.ConnTimeout > defaultConnTimeout {
		s.ConnTimeout = defaultConnTimeout
	}
	return nil
}

//after get socket, the default keep live is 15 seconds, set by "net"
func (s *Socket) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	connClient := &ConnClient{
		Conn:         conn,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		ConnTimeout:  s.ConnTimeout,
	}
	s.ConnClientPool <- connClient
	return err
}

//after get socket, the default keep live is 15 seconds, set by "net"
func (s *Socket) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Listener = ln
	return err
}

func (s *Socket) Close() error {
	for {
		select {
		case conn := <-s.ConnClientPool:
			conn.Close()
		default:
			return nil
		}
	}
}
