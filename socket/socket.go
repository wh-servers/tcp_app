package socket

import (
	"net"
	"time"
)

type Socket struct {
	ConnClientSingle ConnClient
	ConnClientPool   []ConnClient
	Listener         net.Listener
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
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
	return nil
}

//after get socket, the default keep live is 15 seconds, set by "net"
func (s *Socket) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	s.ConnClientSingle = ConnClient{
		Conn: conn,
	}
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
	return s.ConnClientSingle.Close()
}
