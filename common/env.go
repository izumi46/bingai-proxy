package common

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	_        = godotenv.Load()
	HOSTNAME string
	PORT     string
	// user token
	USER_Token string
	// USer Cookie
	USER_KievRPSSecAuth string
	USER_RwBf           string
	USER_MUID           string

	User_Agent        string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"
	User_Agent_Mobile string = "Mozilla/5.0 (iPhone; CPU iPhone OS 15_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.7 Mobile/15E148 Safari/605.1.15 BingSapphire/1.0.410529013"

	LOG_LEVEL = "INFO"
)

func init() {
	HOSTNAME = os.Getenv("HOSTNAME")
	if HOSTNAME == "" {
		HOSTNAME = "localhost"
	}
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	// KievRPSSecAuth Cookie
	USER_KievRPSSecAuth = os.Getenv("USER_KievRPSSecAuth")
	// MUID Cookie
	USER_MUID = os.Getenv("USER_MUID")
	// _RwBf Cookie
	USER_RwBf = os.Getenv("USER_RwBf")
	USER_Token = os.Getenv("USER_Token")

	LOG_LEVEL = strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if LOG_LEVEL == "" {
		LOG_LEVEL = "INFO"
	}
	Logger = NewLogger(LOG_LEVEL)
}
