package main

import (
	"flag"
	"net/http"

	"github.com/brucewang585/go-zero/core/logx"
	"github.com/brucewang585/go-zero/core/service"
	"github.com/brucewang585/go-zero/rest"
	"github.com/brucewang585/go-zero/rest/httpx"
)

var (
	port    = flag.Int("port", 3333, "the port to listen")
	timeout = flag.Int64("timeout", 0, "timeout of milliseconds")
)

type Request struct {
	User string `form:"user,options=a|b"`
}

func first(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Middleware", "first")
		next(w, r)
	}
}

func second(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Middleware", "second")
		next(w, r)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := httpx.Parse(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httpx.OkJson(w, "helllo, "+req.User)
}

func main() {
	flag.Parse()

	engine := rest.MustNewServer(rest.RestConf{
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				Mode: "console",
			},
		},
		Port:     *port,
		Timeout:  *timeout,
		MaxConns: 500,
	}, rest.WithNotAllowedHandler(rest.CorsHandler()))
	defer engine.Stop()

	engine.Use(first)
	engine.Use(second)
	engine.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: handle,
	})
	engine.Start()
}
