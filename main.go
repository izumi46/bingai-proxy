package main

import (
	"adams549659584/go-proxy-bingai/api"
	"adams549659584/go-proxy-bingai/common"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/v1/chat/completions", api.ChatHandler)
	http.HandleFunc("/v1/images/generations", api.ImageHandler)

	addr := common.HOSTNAME + ":" + common.PORT

	common.Logger.Info("Starting BingAI Proxy At " + addr)

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  15 * time.Second,
	}
	common.Logger.Fatal(srv.ListenAndServe())
}
