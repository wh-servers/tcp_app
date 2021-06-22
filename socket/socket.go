package socket

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var defaultPoolSize int32 = 16
var maxPoolSize int32 = 1024
var defaultReadTimeout = 10 * time.Second
var maxReadTimeout = 60 * time.Second
var defaultConnTimeout = 60 * time.Second
var maxConnTimeout = 24 * 60 * 60 * time.Second
var defaultKeepAlivePeriod = 30 * time.Second
var maxKeepAlivePeriod = 60 * time.Second

type Socket struct {
	//ConnClientPool usage: can use load balancer to keep connection live
	ConnClientPool  chan *ConnClient
	PoolSize        int32
	Listener        *net.TCPListener
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ConnTimeout     time.Duration
	KeepAlive       bool
	KeepAlivePeriod time.Duration
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
	if s.PoolSize <= 0 || s.PoolSize > maxPoolSize {
		s.PoolSize = defaultPoolSize
	}
	s.ConnClientPool = make(chan *ConnClient, s.PoolSize)
	//set conn timeout
	if s.ReadTimeout <= 0 || s.ReadTimeout > maxReadTimeout {
		s.ReadTimeout = defaultReadTimeout
	}
	if s.ConnTimeout <= 0 || s.ConnTimeout > maxConnTimeout {
		s.ConnTimeout = defaultConnTimeout
	}
	//set keep alive
	if s.KeepAlivePeriod <= 0 || s.KeepAlivePeriod > maxKeepAlivePeriod {
		s.KeepAlivePeriod = defaultKeepAlivePeriod
	}
	return nil
}

//after get socket, the default keep live is 15 seconds, set by "net"
func (s *Socket) Dial(addr string) error {
	strs := strings.Split(addr, ":")
	port, err := strconv.Atoi(strs[1])
	if err != nil || len(strs) < 2 {
		return fmt.Errorf("addr format err")
	}
	ipStrArr := strings.Split(strs[0], ".")
	byteArr := []byte{}
	var ip net.IP
	switch {
	case len(ipStrArr) == 0:
		break
	case len(ipStrArr) == 4:
		for _, b := range ipStrArr {
			bInt, err := strconv.Atoi(b)
			if err != nil {
				return fmt.Errorf("addr format err")
			}
			byteArr = append(byteArr, byte(bInt))
		}
		ip = net.IPv4(byteArr[0], byteArr[1], byteArr[2], byteArr[3])
	}
	tcpAddr := &net.TCPAddr{
		IP:   ip,
		Port: port,
		Zone: "",
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	connClient := NewConnClient()
	err = connClient.Init(s, conn)
	if err != nil {
		return err
	}
	s.ConnClientPool <- connClient
	return err
}

//after get socket, the default keep live is 15 seconds, set by "net"
func (s *Socket) Listen(addr string) error {
	strs := strings.Split(addr, ":")
	port, err := strconv.Atoi(strs[1])
	if err != nil || len(strs) < 2 {
		return fmt.Errorf("addr format err")
	}
	tcpAddr := &net.TCPAddr{
		IP:   nil,
		Port: port,
		Zone: "",
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
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
