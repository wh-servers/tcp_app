package socket

import (
	"time"
)

type Option interface {
	Apply(s *Socket) error
}

type ReadTimeoutOption struct {
	ReadTimeout int32
}

type WriteTimeoutOption struct {
	WriteTimeout int32
}

type ConnectionOption struct {
	PoolSize int32
}

type KeepAliveOption struct {
	KeepAlive       bool
	KeepAlivePeriod int32
}

func (r *ReadTimeoutOption) Apply(s *Socket) error {
	s.ReadTimeout = time.Duration(r.ReadTimeout) * time.Second
	return nil
}

func (w *WriteTimeoutOption) Apply(s *Socket) error {
	s.WriteTimeout = time.Duration(w.WriteTimeout) * time.Second
	return nil
}

func (c *ConnectionOption) Apply(s *Socket) error {
	s.PoolSize = c.PoolSize
	return nil
}

func (k *KeepAliveOption) Apply(s *Socket) error {
	s.KeepAlive = k.KeepAlive
	s.KeepAlivePeriod = time.Duration(k.KeepAlivePeriod) * time.Second
	return nil
}
