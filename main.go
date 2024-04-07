package main

import (
	"izumi46/bingai-proxy/api"
	"izumi46/bingai-proxy/common"
	"net"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/v1/chat/completions", api.ChatHandler)
	http.HandleFunc("/v1/images/generations", api.ImageHandler)

	if common.DOMAIN_SOCKET == "" {
		addr := common.HOSTNAME + ":" + common.PORT

		common.Logger.Info("Starting BingAI Proxy At " + addr)

		srv := &http.Server{
			Addr:         addr,
			WriteTimeout: 5 * time.Minute,
			ReadTimeout:  15 * time.Second,
		}
		common.Logger.Fatal(srv.ListenAndServe())
	} else {
		path := common.DOMAIN_SOCKET
		common.Logger.Info("Starting BingAI Proxy At " + path)
		srv := &http.Server{}
		unixListener, err := net.Listen("unix", path)
		if err != nil {
			panic(err)
		}
		common.Logger.Fatal(srv.Serve(unixListener))
	}
}
