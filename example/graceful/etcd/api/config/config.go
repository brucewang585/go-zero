package config

import (
	"github.com/brucewang585/go-zero/rest"
	"github.com/brucewang585/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Rpc zrpc.RpcClientConf
}
