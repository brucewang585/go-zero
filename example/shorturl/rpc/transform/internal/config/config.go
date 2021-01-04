package config

import (
	"github.com/brucewang585/go-zero/core/stores/cache"
	"github.com/brucewang585/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Table      string
	Cache      cache.CacheConf
}
