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
	Conn net.Conn
	//todo: timeout is not in use yet
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

//first 4 bytes are to indicate the main msg length. the length except these 4 bytes
//the 5th byte is the command number, used to tell which handler to use in server side.
func (c *ConnClient) Read(msg *[]byte) error {
	var resLen int32
	if c == nil || c.Conn == nil {
		return fmt.Errorf("nil connection")
	}
	//read main msg length
	err := binary.Read(c.Conn, binary.LittleEndian, &resLen)
	if err != nil {
		return err
	}
	//read main msg
	buf := make([]byte, resLen)
	_, err = io.ReadFull(c.Conn, buf)
	if err != nil {
		return fmt.Errorf("read from conn error: ", err)
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
