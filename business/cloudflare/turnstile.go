package cloudflare

const SECRET_KEY = "1x0000000000000000000000000000000AA"

func VerifyTurnstileToken(token string) bool {
	return token == SECRET_KEY
}
