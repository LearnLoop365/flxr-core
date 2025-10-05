package sec

type AuthRequestBody struct {
	AuthClientID string `json:"auth_client_id"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"` // Required for Security Check. RFC 6749 §4.1.3
}

type ReissueAccessTokenRequestBody struct {
	RefreshToken string `json:"refresh_token"`
	UID          int    `json:"uid"` // for Checking
}
