package dto

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Registration struct {
	Username string `json:"username"`
}

type RefreshToken struct {
	ID             string `json:"id"`
	Token          string `json:"jwt"`
	ExpireDuration int    `json:"expire_duration"`
}

type AccessToken struct {
	Username       string `json:"username"`
	RefreshTokenId string `json:"refresh_token_id"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
