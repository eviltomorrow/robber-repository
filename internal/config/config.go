package config

import (
	"encoding/json"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Log    Log    `json:"log" toml:"log"`
	MySQL  MySQL  `json:"mysql" toml:"mysql"`
	Etcd   Etcd   `json:"etcd" toml:"etcd"`
	Server Server `json:"server" toml:"server"`
}

type Log struct {
	DisableTimestamp bool   `json:"disable-timestamp" toml:"disable-timestamp"`
	Level            string `json:"level" toml:"level"`
	Format           string `json:"format" toml:"format"`
	FileName         string `json:"filename" toml:"filename"`
	MaxSize          int    `json:"maxsize" toml:"maxsize"`
}

type MySQL struct {
	DSN     string `json:"dsn" toml:"dsn"`
	MinOpen int    `json:"min-open" toml:"min-open"`
	MaxOpen int    `json:"max-open" toml:"max-open"`
}

type Etcd struct {
	Endpoints []string `json:"endpoints" toml:"endpoints"`
}

type Server struct {
	Host string `json:"host" toml:"host"`
	Port int    `json:"port" toml:"port"`
}

func (c *Config) Load(path string, override func(cfg *Config)) error {
	if path == "" {
		return nil
	}

	if _, err := toml.DecodeFile(path, c); err != nil {
		return err
	}
	return nil
}

func (cg *Config) String() string {
	buf, _ := json.Marshal(cg)
	return string(buf)
}

var GlobalConfig = &Config{
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		FileName:         "/tmp/robber-repository/data.log",
		MaxSize:          20,
	},
	MySQL: MySQL{
		DSN:     "root:root@tcp(127.0.0.1:3306)/robber?charset=utf8mb4&parseTime=true&loc=Local",
		MinOpen: 5,
		MaxOpen: 10,
	},
	Etcd: Etcd{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
	},
	Server: Server{
		Host: "0.0.0.0",
		Port: 27321,
	},
}
