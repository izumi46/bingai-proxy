package api

import (
	"izumi46/bingai-proxy/common"
)

var Cookie string

func init() {
	go func() {
		Cookie = getCookie()
		common.Logger.Info("BingAPI Ready!")
	}()
}

func getCookie() (cookie string) {
	cookie = ""
	if common.USER_KievRPSSecAuth != "" {
		cookie += "; KievRPSSecAuth=" + common.USER_KievRPSSecAuth
	}
	if common.USER_RwBf != "" {
		cookie += "; _RwBf=" + common.USER_RwBf
	}
	if common.USER_MUID != "" {
		cookie += "; MUID=" + common.USER_MUID
	}
	if common.USER_Token != "" {
		cookie += "; _U=" + common.USER_Token
	}
	return cookie
}
