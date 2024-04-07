package api

import (
	"adams549659584/go-proxy-bingai/bing"
	"adams549659584/go-proxy-bingai/common"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	STOPFLAG = "stop"
)

func generateRandomHex(n int) string {
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	chat := bing.NewChat(Cookie)

	resqB, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		common.Logger.Error("ReadAll Error: %v", err)
		return
	}

	var resq chatRequest
	err = json.Unmarshal(resqB, &resq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		common.Logger.Error("Unmarshal Error: %v", err)
		return
	}

	err = chat.NewConversation()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		common.Logger.Error("NewConversation Error: %v", err)
		return
	}

	chat.SetStyle(strings.ReplaceAll(resq.Model, "-vision", ""))

	prompt, msg, image := chat.MsgComposer(resq.Messages)

	if !strings.Contains(resq.Model, "-vision") && image != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Image is not supported in this model"))
		common.Logger.Error("Chat Error: Image is not supported in this model")
		return
	}

	resp := chatResponse{
		Id:                "chatcmpl-NewBing",
		Object:            "chat.completion.chunk",
		SystemFingerprint: generateRandomHex(12),
		Model:             resq.Model,
		Create:            time.Now().Unix(),
	}

	if resq.Stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")

		text := make(chan string)
		defer close(text)
		go chat.ChatStream(prompt, msg, text, image)
		var tmp string

		for {
			tmp = <-text
			resp.Choices = []choices{
				{
					Index: 0,
					Delta: bing.Message{
						// Role:    "assistant",
						Content: tmp,
					},
				},
			}
			if tmp == "EOF" {
				resp.Choices[0].Delta.Content = ""
				resp.Choices[0].FinishReason = &STOPFLAG
				resData, err := json.Marshal(resp)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					common.Logger.Error("Marshal Error: %v", err)
					return
				}
				w.Write([]byte("data: "))
				w.Write(resData)
				w.Write([]byte("\n\n"))
				flusher.Flush()
				break
			}
			resData, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				common.Logger.Error("Marshal Error: %v", err)
				return
			}
			w.Write([]byte("data: "))
			w.Write(resData)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
		w.Write([]byte("data: [DONE]\n"))
		flusher.Flush()
	} else {
		text, err := chat.Chat(prompt, msg, image)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			common.Logger.Error("Chat Error: %v", err)
			return
		}

		resp.Choices = append(resp.Choices, choices{
			Index: 0,
			Message: bing.Message{
				Role:    "assistant",
				Content: text,
			},
			FinishReason: &STOPFLAG,
		})

		resData, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			common.Logger.Error("Marshal Error: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resData)
	}
}
