package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/brucewang585/go-zero/core/executors"
	"github.com/brucewang585/go-zero/core/logx"
	"github.com/brucewang585/go-zero/example/graceful/dns/api/svc"
	"github.com/brucewang585/go-zero/example/graceful/dns/api/types"
	"github.com/brucewang585/go-zero/example/graceful/dns/rpc/graceful"
	"github.com/brucewang585/go-zero/rest/httpx"
)

func gracefulHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	logger := executors.NewLessExecutor(time.Second)
	return func(w http.ResponseWriter, r *http.Request) {
		host, err := os.Hostname()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}

		conn := ctx.Client.Conn()
		client := graceful.NewGraceServiceClient(conn)
		rp, err := client.Grace(context.Background(), &graceful.Request{From: host})
		if err != nil {
			logx.Error(err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		var resp types.Response
		resp.Host = rp.Host
		logger.DoOrDiscard(func() {
			fmt.Printf("%s from host: %s\n", time.Now().Format("15:04:05"), rp.Host)
		})
		httpx.OkJson(w, resp)
	}
}
