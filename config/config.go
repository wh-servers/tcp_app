package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/wh-servers/tcp_app/socket"
	"gopkg.in/yaml.v2"
)

type Config struct {
	//todo: how to use socket option
	SocketOptions []socket.Option
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config (path=%s) error: %v", path, err)
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read config (path=%s) error: %v", path, err)
	}
	if err := yaml.Unmarshal(bs, c); err != nil {
		return fmt.Errorf("unmarshal config (path=%s) error: %v", path, err)
	}
	return nil

}
