package socket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type ConnClient struct {
	Conn         *net.TCPConn
	IsDead       chan bool
	ConnTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewConnClient() *ConnClient {
	return &ConnClient{}
}

func (c *ConnClient) Init(s *Socket, conn *net.TCPConn) error {
	c.Conn = conn
	c.ReadTimeout = s.ReadTimeout
	c.WriteTimeout = s.WriteTimeout
	c.ConnTimeout = s.ConnTimeout
	c.IsDead = make(chan bool, 1)
	//set keep alive
	//err := conn.SetKeepAlive(s.KeepAlive)
	err := conn.SetKeepAlive(true)
	if err != nil {
		return err
	}
	//e.g. conn break after: 5 sec + 8 * 5 sec
	//err = conn.SetKeepAlivePeriod(s.KeepAlivePeriod)
	err = conn.SetKeepAlivePeriod(2 * time.Second)
	return err
}

//first 4 bytes are to indicate the main msg length. the length except these 4 bytes
//the 5th byte is the command number, used to tell which handler to use in server side.
func (c *ConnClient) Read(msg *[]byte) error {
	var resLen int32
	if c == nil || c.Conn == nil {
		return fmt.Errorf("nil connection")
	}
	err := c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return fmt.Errorf("SetReadDeadline err: %v", err)
	}
	//read main msg length
	err = binary.Read(c.Conn, binary.LittleEndian, &resLen)
	fmt.Println("debuggg: msg", resLen)
	if err != nil {
		//if err == io.EOF {
		//	c.IsDead <- true
		//}
		return err
	}
	//read main msg
	if resLen < 0 {
		resLen = 0
	}
	buf := make([]byte, resLen)
	_, err = io.ReadFull(c.Conn, buf)
	if err != nil {
		return fmt.Errorf("read from conn error: %v", err)
	}
	*msg = buf
	return err
}

func (c *ConnClient) Write(msg []byte) error {
	if c == nil || c.Conn == nil {
		return fmt.Errorf("nil connection")
	}
	buf := bytes.Buffer{}
	//write msg length value to buf, it is an int32 in byte format, byte order by LitterEndian
	err := binary.Write(&buf, binary.LittleEndian, int32(len(msg)))
	if err != nil {
		return err
	}
	//write main msg to buf
	_, err = buf.Write(msg)
	if err != nil {
		return err
	}
	//write buf data into conn
	_, err = buf.WriteTo(c.Conn)
	if err != nil {
		return err
	}
	return err
}

func (c *ConnClient) Close() error {
	return c.Conn.Close()
}
