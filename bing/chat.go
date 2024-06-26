package bing

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	PRECISE          = "Precise"
	BALANCED         = "Balanced"
	CREATIVE         = "Creative"
	PRECISE_OFFLINE  = "Precise-offline"
	BALANCED_OFFLINE = "Balanced-offline"
	CREATIVE_OFFLINE = "Creative-offline"

	PRECISE_G4T          = "Precise-g4t"
	BALANCED_G4T         = "Balanced-g4t"
	CREATIVE_G4T         = "Creative-g4t"
	PRECISE_G4T_OFFLINE  = "Precise-g4t-offline"
	BALANCED_G4T_OFFLINE = "Balanced-g4t-offline"
	CREATIVE_G4T_OFFLINE = "Creative-g4t-offline"

	PRECISE_18K          = "Precise-18k"
	BALANCED_18K         = "Balanced-18k"
	CREATIVE_18K         = "Creative-18k"
	PRECISE_18K_OFFLINE  = "Precise-18k-offline"
	BALANCED_18K_OFFLINE = "Balanced-18k-offline"
	CREATIVE_18K_OFFLINE = "Creative-18k-offline"

	PRECISE_G4T_18K  = "Precise-g4t-18k"
	BALANCED_G4T_18K = "Balanced-g4t-18k"
	CREATIVE_G4T_18K = "Creative-g4t-18k"
)

var ChatModels = [21]string{BALANCED, BALANCED_OFFLINE, CREATIVE, CREATIVE_OFFLINE, PRECISE, PRECISE_OFFLINE, BALANCED_G4T, BALANCED_G4T_OFFLINE, CREATIVE_G4T, CREATIVE_G4T_OFFLINE, PRECISE_G4T, PRECISE_G4T_OFFLINE,
	BALANCED_18K, BALANCED_18K_OFFLINE, CREATIVE_18K, CREATIVE_18K_OFFLINE, PRECISE_18K, PRECISE_18K_OFFLINE, BALANCED_G4T_18K, CREATIVE_G4T_18K, PRECISE_G4T_18K}

const (
	bingCreateConversationUrl = "%s/turing/conversation/create?bundleVersion=1.1467.6"
	sydneyChatHubUrl          = "%s/sydney/ChatHub?sec_access_token=%s"
	imagesKblob               = "%s/images/kblob"
	imageUploadUrl            = "%s/images/blob?bcid=%s"

	spilt = "\x1e"
)

func generateRandomHex(n int) string {
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func NewChat(cookies string) *Chat {
	return &Chat{
		cookies:       cookies,
		BingBaseUrl:   bingBaseUrl,
		SydneyBaseUrl: sydneyBaseUrl,
	}
}

func (chat *Chat) SetStyle(style string) *Chat {
	chat.GetChatHub().SetStyle(style)
	return chat
}

func (chat *Chat) GetChatHub() *ChatHub {
	return chat.chatHub
}

func (chat *Chat) GetStyle() string {
	return chat.GetChatHub().GetStyle()
}

func (chat *Chat) GetTone() string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(chat.GetStyle(), "-18k", ""), "-g4t", ""), "-offline", "")
}

func (chat *Chat) NewConversation() error {
	URL := fmt.Sprintf(bingCreateConversationUrl, chat.BingBaseUrl)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Host", "www.bing.com")
	req.Header.Set("Origin", "https://www.bing.com")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", chat.cookies)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US;q=0.9")
	req.Header.Set("Referer", "https://www.bing.com/chat?q=Bing+AI&showconv=1&FORM=hpcodx")
	req.Header.Set("Sec-Ch-Ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Microsoft Edge\";v=\"120\"")
	req.Header.Set("Sec-Ch-Ua-Arch", "\"x86\"")
	req.Header.Set("Sec-Ch-Ua-Bitness", "\"64\"")
	req.Header.Set("Sec-Ch-Ua-Full-Version", "\"120.0.2210.133\"")
	req.Header.Set("Sec-Ch-Ua-Full-Version-List", "\"Not_A Brand\";v=\"8.0.0.0\", \"Chromium\";v=\"120.0.6099.217\", \"Microsoft Edge\";v=\"120.0.2210.133\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Model", "\"\"")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req.Header.Set("Sec-Ch-Ua-Platform-Version", "\"15.0.0\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Ms-Useragent", "azsdk-js-api-client-factory/1.0.0-beta.1 core-rest-pipeline/1.12.3 OS/Windows")
	req.Header.Set("X-Ms-Client-Request-Id", uuid.NewString())
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var chatReq ChatReq
	err = json.Unmarshal(bytes, &chatReq)
	if err != nil {
		return err
	}
	chatReq.ConversationSignature = res.Header.Get("X-Sydney-Conversationsignature")
	chatReq.EncryptedConversationSignature = res.Header.Get("X-Sydney-Encryptedconversationsignature")

	chat.chatHub = newChatHub(chatReq)

	return nil
}

func (chat *Chat) MsgComposer(msgs []Message) (prompt string, msg string, image string) {
	systemMsgNum := 0
	for index, t := range msgs {
		if t.Role == "system" {
			systemMsgNum++
			switch t.Content.(type) {
			case string:
				prompt = t.Content.(string)
			case []interface{}:
				for _, v := range t.Content.([]interface{}) {
					value := v.(map[string]interface{})
					if strings.ToLower(value["type"].(string)) == "text" {
						prompt = value["text"].(string)
					}
				}
			case []ContentPart:
				for _, v := range t.Content.([]ContentPart) {
					if strings.ToLower(v.Type) == "text" {
						prompt = v.Text
					}
				}
			default:
				continue
			}
			msgs = append(msgs[0:index], msgs[index+1:]...)
		}
	}
	if len(msgs)-systemMsgNum == 1 {
		switch msgs[0].Content.(type) {
		case string:
			return prompt, msgs[0].Content.(string), ""
		case []interface{}:
			tmp := ""
			for _, v := range msgs[0].Content.([]interface{}) {
				value := v.(map[string]interface{})
				if strings.ToLower(value["type"].(string)) == "text" {
					tmp += value["text"].(string)
				} else if strings.ToLower(value["type"].(string)) == "image_url" {
					image = value["image_url"].(map[string]interface{})["url"].(string)
				}
			}
			return prompt, tmp, image
		case []ContentPart:
			tmp := ""
			for _, v := range msgs[0].Content.([]ContentPart) {
				if strings.ToLower(v.Type) == "text" {
					tmp += v.Text
				} else if strings.ToLower(v.Type) == "image_url" {
					image = v.ImageUrl.Url
				}
			}
			return prompt, tmp, image
		default:
			return prompt, "", ""
		}
	}

	var lastRole string
	for _, t := range msgs {
		tmp := ""
		switch t.Content.(type) {
		case string:
			tmp = t.Content.(string)
		default:
			tmp = ""
			for _, v := range msgs[0].Content.([]ContentPart) {
				if strings.ToLower(v.Type) == "text" {
					tmp += v.Text
				} else if strings.ToLower(v.Type) == "image_url" {
					image = v.ImageUrl.Url
				}
			}
		}
		if lastRole == t.Role {
			msg += "\n" + tmp
			continue
		} else if lastRole != "" {
			msg += "\n\n"
		}
		switch t.Role {
		case "system":
			prompt += tmp
		case "user":
			msg += "`me`:\n" + tmp
		case "assistant":
			msg += "`you`:\n" + tmp
		}
		if t.Role != "system" {
			lastRole = t.Role
		}
	}
	msg += "\n\n`you`:"
	return prompt, msg, image
}

func (chat *Chat) optionsSetsHandler(systemContext []SystemContext) []string {
	optionsSets := []string{
		"nlu_direct_response_filter",
		"deepleo",
		"disable_emoji_spoken_text",
		"responsible_ai_policy_235",
		"enablemm",
		"dv3sugg",
		"iyxapbing",
		"iycapbing",
		"enable_user_consent",
		"fluxmemcst",
		"gldcl1p",
		"uquopt",
		"langdtwb",
		"enflst",
		"enpcktrk",
		"rcaldictans",
		"rcaltimeans",
		"gndbfptlw",
	}
	if len(systemContext) > 0 {
		optionsSets = append(optionsSets, "nojbfedge", "rai278")
	}

	tone := chat.GetStyle()
	if strings.Contains(tone, "g4t") {
		optionsSets = append(optionsSets, "dlgpt4t")
	}
	if strings.Contains(tone, "18k") {
		optionsSets = append(optionsSets, "prjupy")
	}
	if strings.Contains(tone, PRECISE) {
		optionsSets = append(optionsSets, "h3precise", "clgalileo", "gencontentv3")
	} else if strings.Contains(tone, BALANCED) {
		if strings.Contains(tone, "18k") {
			optionsSets = append(optionsSets, "clgalileo", "saharagenconv5")
		} else {
			optionsSets = append(optionsSets, "galileo", "saharagenconv5")
		}
	} else if strings.Contains(tone, CREATIVE) {
		optionsSets = append(optionsSets, "h3imaginative", "clgalileo", "gencontentv3")
	}
	return optionsSets
}

func (chat *Chat) sliceIdsHandler(systemContext []SystemContext) []string {
	if len(systemContext) > 0 {
		return []string{
			"winmuid1tf",
			"styleoff",
			"ccadesk",
			"smsrpsuppv4cf",
			"ssrrcache",
			"contansperf",
			"crchatrev",
			"winstmsg2tf",
			"creatgoglt",
			"creatorv2t",
			"sydconfigoptt",
			"adssqovroff",
			"530pstho",
			"517opinion",
			"418dhlth",
			"512sprtic1s0",
			"emsgpr",
			"525ptrcps0",
			"529rweas0",
			"515oscfing2s0",
			"524vidansgs0",
		}
	} else {
		return []string{
			"qnacnc",
			"fluxsunoall",
			"mobfdbk",
			"v6voice",
			"cmcallcf",
			"specedge",
			"tts5",
			"advperfcon",
			"designer2cf",
			"defred",
			"msgchkcf",
			"thrdnav",
			"0212boptpsc",
			"116langwb",
			"124multi2t",
			"927storev2s0",
			"0131dv1",
			"1pgptwdes",
			"0131gndbfpr",
			"brndupdtcf",
			"enter4nl",
		}
	}
}

func (chat *Chat) pluginHandler(optionsSets *[]string) []Plugins {
	plugins := []Plugins{}
	tone := chat.GetStyle()
	if !strings.Contains(tone, "offline") {
		plugins = append(plugins, Plugins{Id: "c310c353-b9f0-4d76-ab0d-1dd5e979cf68"})
	} else {
		*optionsSets = append(*optionsSets, "nosearchall")
	}
	return plugins
}

func (chat *Chat) systemContextHandler(prompt string) []SystemContext {
	systemContext := []SystemContext{}
	if prompt != "" {
		systemContext = append(systemContext, SystemContext{
			Author:      "user",
			Description: prompt,
			ContextType: "WebPage",
			MessageType: "Context",
			// MessageId:   "discover-web--page-ping-mriduna-----",
			SourceName: "Ubuntu Pastebin",
			SourceUrl:  "https://paste.ubuntu.com/p/" + generateRandomHex(10) + "/",
		})
	}
	return systemContext
}

func (chat *Chat) requestPayloadHandler(msg string, optionsSets []string, sliceIds []string, plugins []Plugins, systemContext []SystemContext, imageUrl string) (data map[string]any, msgId string) {
	msgId = uuid.NewString()

	data = map[string]any{
		"arguments": []any{
			map[string]any{
				"source":      "cib",
				"optionsSets": optionsSets,
				"allowedMessageTypes": []string{
					"ActionRequest",
					"Chat",
					"ConfirmationCard",
					"Context",
					"InternalSearchQuery",
					"InternalSearchResult",
					"Disengaged",
					"InternalLoaderMessage",
					"InvokeAction",
					"Progress",
					"RenderCardRequest",
					"RenderContentRequest",
					"AdsQuery",
					"SemanticSerp",
					"GenerateContentQuery",
					"SearchQuery",
				},
				"sliceIds":         sliceIds,
				"isStartOfSession": true,
				"gptId":            "copilot",
				"verbosity":        "verbose",
				"scenario":         "SERP",
				"plugins":          plugins,
				"previousMessages": systemContext,
				"traceId":          strings.ReplaceAll(uuid.NewString(), "-", ""),
				"conversationHistoryOptionsSets": []string{
					"autosave",
					"savemem",
					"uprofupd",
					"uprofgen",
				},
				"requestId": msgId,
				"message":   chat.requestMessagePayloadHandler(msg, msgId, imageUrl),
				// "conversationSignature": chat.GetChatHub().GetConversationSignature(),
				"tone":           chat.GetTone(),
				"spokenTextMode": "None",
				"participant": map[string]any{
					"id": chat.GetChatHub().GetClientId(),
				},
				"conversationId": chat.GetChatHub().GetConversationId(),
			},
		},
		"invocationId": "0",
		"target":       "chat",
		"type":         4,
	}

	return
}

func (chat *Chat) requestMessagePayloadHandler(msg string, msgId string, imageUrl string) map[string]any {
	if imageUrl != "" {
		return map[string]any{
			"author":           "user",
			"inputMethod":      "Keyboard",
			"imageUrl":         imageUrl,
			"originalImageUrl": imageUrl,
			"text":             msg,
			"messageType":      "Chat",
			"requestId":        msgId,
			"messageId":        msgId,
			// "userIpAddress":    "1.1.1.1",
			"locale":   "zh-CN",
			"market":   "en-US",
			"region":   "CN",
			"location": "lat:47.639557;long:-122.128159;re=1000m;",
			"locationHints": []any{
				map[string]any{
					"country":           "United States",
					"state":             "California",
					"city":              "Los Angeles",
					"timezoneoffset":    8,
					"dma":               819,
					"countryConfidence": 8,
					"cityConfidence":    5,
					"Center": map[string]any{
						"Latitude":  78.4156,
						"Longitude": -101.4458,
					},
					"RegionType": 2,
					"SourceType": 1,
				},
			},
		}
	}

	return map[string]any{
		"author":      "user",
		"inputMethod": "Keyboard",
		"text":        msg,
		"messageType": "Chat",
		"requestId":   msgId,
		"messageId":   msgId,
		// "userIpAddress": "1.1.1.1",
		"locale":   "zh-CN",
		"market":   "en-US",
		"region":   "CN",
		"location": "lat:47.639557;long:-122.128159;re=1000m;",
		"locationHints": []any{
			map[string]any{
				"country":           "United States",
				"state":             "California",
				"city":              "Los Angeles",
				"timezoneoffset":    8,
				"dma":               819,
				"countryConfidence": 8,
				"cityConfidence":    5,
				"Center": map[string]any{
					"Latitude":  78.4156,
					"Longitude": -101.4458,
				},
				"RegionType": 2,
				"SourceType": 1,
			},
		},
	}
}

func (chat *Chat) imageUploadHandler(image string) (string, error) {
	if strings.HasPrefix(image, "http") {
		return image, nil
	}
	if strings.Contains(image, "base64,") {
		image = strings.Split(image, ",")[1]
		buf := new(bytes.Buffer)
		bw := multipart.NewWriter(buf)
		p1, _ := bw.CreateFormField("knowledgeRequest")
		p1.Write([]byte(fmt.Sprintf("{\"imageInfo\":{},\"knowledgeRequest\":{\"invokedSkills\":[\"ImageById\"],\"subscriptionId\":\"Bing.Chat.Multimodal\",\"invokedSkillsRequestData\":{\"enableFaceBlur\":true},\"convoData\":{\"convoid\":\"%s\",\"convotone\":\"%s\"}}}", chat.GetChatHub().GetConversationId(), chat.GetTone())))
		p2, _ := bw.CreateFormField("imageBase64")
		p2.Write([]byte(strings.ReplaceAll(image, " ", "+")))
		bw.Close()

		URL := fmt.Sprintf(imagesKblob, chat.BingBaseUrl)
		req, err := http.NewRequest("POST", URL, buf)
		if err != nil {
			return "", err
		}
		req.Header.Set("Host", "www.bing.com")
		req.Header.Set("Origin", "https://www.bing.com")
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Cookie", chat.cookies)
		req.Header.Set("Content-Type", "multipart/form-data")
		req.Header.Set("Referer", "https://www.bing.com/chat?q=Bing+AI&showconv=1&FORM=hpcodx")
		req.Header.Set("Sec-Ch-Ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Microsoft Edge\";v=\"120\"")
		req.Header.Set("Sec-Ch-Ua-Arch", "\"x86\"")
		req.Header.Set("Sec-Ch-Ua-Bitness", "\"64\"")
		req.Header.Set("Sec-Ch-Ua-Full-Version", "\"120.0.2210.133\"")
		req.Header.Set("Sec-Ch-Ua-Full-Version-List", "\"Not_A Brand\";v=\"8.0.0.0\", \"Chromium\";v=\"120.0.6099.217\", \"Microsoft Edge\";v=\"120.0.2210.133\"")
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Model", "\"\"")
		req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
		req.Header.Set("Sec-Ch-Ua-Platform-Version", "\"15.0.0\"")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		var image imageUploadStruct
		err = json.Unmarshal(bytes, &image)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(imageUploadUrl, chat.BingBaseUrl, image.BlobId), nil
	}
	return "", nil
}

func (chat *Chat) wsHandler(data map[string]any) (*websocket.Conn, error) {
	dialer := websocket.DefaultDialer
	dialer.Proxy = http.ProxyFromEnvironment
	headers := http.Header{}
	headers.Set("Accept-Encoding", "gzip, deflate, br")
	headers.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Pragma", "no-cache")
	headers.Set("User-Agent", userAgent)
	headers.Set("Referer", "https://www.bing.com/chat?q=Bing+AI&showconv=1&FORM=hpcodx")
	headers.Set("Cookie", chat.cookies)
	headers.Set("Host", "sydney.bing.com")
	headers.Set("Origin", "https://www.bing.com")

	ws, _, err := dialer.Dial(fmt.Sprintf(sydneyChatHubUrl, chat.SydneyBaseUrl, url.QueryEscape(chat.GetChatHub().GetEncryptedConversationSignature())), headers)
	if err != nil {
		return nil, err
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("{\"protocol\":\"json\",\"version\":1}"+spilt))
	if err != nil {
		return nil, err
	}

	_, _, err = ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":6}"+spilt))
	if err != nil {
		return nil, err
	}

	req, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = ws.WriteMessage(websocket.TextMessage, append(req, []byte(spilt)...))
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (chat *Chat) Chat(prompt, msg string, image ...string) (string, error) {
	c := make(chan string)
	go func() {
		tmp := ""
		for {
			tmp = <-c
			if tmp == "EOF" {
				break
			}
		}
	}()

	return chat.chatHandler(prompt, msg, c, image...)
}

func (chat *Chat) ChatStream(prompt, msg string, c chan string, image ...string) (string, error) {
	return chat.chatHandler(prompt, msg, c, image...)
}

func (chat *Chat) chatHandler(prompt, msg string, c chan string, image ...string) (string, error) {
	imageUrl := ""
	if len(image) > 0 {
		url, err := chat.imageUploadHandler(image[0])
		if err != nil {
			c <- "EOF"
			return "", err
		}
		imageUrl = url
	}
	systemContext := chat.systemContextHandler(prompt)
	optionsSets := chat.optionsSetsHandler(systemContext)
	sliceIds := chat.sliceIdsHandler(systemContext)
	plugins := chat.pluginHandler(&optionsSets)
	data, _ := chat.requestPayloadHandler(msg, optionsSets, sliceIds, plugins, systemContext, imageUrl)

	ws, err := chat.wsHandler(data)
	if err != nil {
		c <- "EOF"
		return "", err
	}
	defer ws.Close()

	text := ""

	i := 0
	for {
		if i >= 15 {
			err := ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":6}"+spilt))
			if err != nil {
				break
			}
			i = 0
		}
		resp := new(ResponsePayload)
		err = ws.ReadJSON(&resp)
		if err != nil {
			if err.Error() != "EOF" {
				c <- "EOF"
				return text, err
			} else {
				c <- "EOF"
				return text, nil
			}
		}
		if resp.Type == 2 {
			if resp.Item.Result.Value == "CaptchaChallenge" || resp.Item.Result.Value == "Throttled" {
				if resp.Item.Result.Value == "CaptchaChallenge" {
					text = "User needs to solve CAPTCHA to continue."
					c <- "User needs to solve CAPTCHA to continue."
				} else if resp.Item.Result.Value == "Throttled" {
					text = "Request is throttled."
					c <- "Request is throttled."
					text = "Unknown error."
				} else {
					c <- "Unknown error."
				}
				break
			} else if resp.Item.Result.Value == "Success" {
				if len(resp.Item.Messages) > 1 {
					for i, v := range resp.Item.Messages[len(resp.Item.Messages)-1].SourceAttributions {
						c <- "\n[^" + strconv.Itoa(i+1) + "^]: [" + v.ProviderDisplayName + "](" + v.SeeMoreUrl + ")"
						text += "\n[^" + strconv.Itoa(i+1) + "^]: [" + v.ProviderDisplayName + "](" + v.SeeMoreUrl + ")"
					}
				}
				break
			}
		} else if resp.Type == 1 {
			if len(resp.Arguments) > 0 {
				if len(resp.Arguments[0].Messages) > 0 {
					if resp.Arguments[0].Messages[0].MessageType == "InternalSearchResult" {
						continue
					}
					if resp.Arguments[0].Messages[0].MessageType == "InternalSearchQuery" || resp.Arguments[0].Messages[0].MessageType == "InternalLoaderMessage" {
						c <- resp.Arguments[0].Messages[0].Text
						c <- "\n\n"
						continue
					}
					if len(resp.Arguments[0].Messages[0].Text) > len(text) {
						c <- strings.ReplaceAll(resp.Arguments[0].Messages[0].Text, text, "")
						text = resp.Arguments[0].Messages[0].Text
					}
					// fmt.Println(resp.Arguments[0].Messages[0].Text + "\n\n")
				}
			}
		} else if resp.Type == 3 {
			break
		} else if resp.Type == 6 {
			err := ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":6}"+spilt))
			if err != nil {
				break
			}
			i = 0
		}
		i++
	}

	c <- "EOF"

	return text, nil
}
