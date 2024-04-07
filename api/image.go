package api

import (
	"encoding/json"
	"io"
	"izumi46/bingai-proxy/bing"
	"izumi46/bingai-proxy/common"
	"net/http"
	"time"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.Header().Add("Allow", "POST")
		w.Header().Add("Access-Control-Allow-Method", "POST")
		w.Header().Add("Access-Control-Allow-Header", "Content-Type, Authorization")
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
		return
	}

	image := bing.NewImage(Cookie)

	resqB, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		common.Logger.Error("ReadAll Error: %v", err)
		return
	}

	var resq imageRequest
	err = json.Unmarshal(resqB, &resq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		common.Logger.Error("Unmarshal Error: %v", err)
		return
	}

	resp := imageResponse{
		Created: time.Now().Unix(),
	}

	if resq.Prompt == "" {
		resData, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			common.Logger.Error("Marshal Error: %v", err)
			return
		}
		w.Write(resData)
		return
	}

	imgs, _, err := image.Image(resq.Prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		common.Logger.Error("Image Error: %v", err)
		return
	}

	for _, img := range imgs {
		resp.Data = append(resp.Data, imageData{
			Url: img,
		})
	}

	resData, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		common.Logger.Error("Marshal Error: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resData)
}
