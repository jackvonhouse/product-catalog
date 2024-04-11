package dto

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Registration struct {
	Username string `json:"username"`
}

type RefreshToken struct {
	ID             string `json:"id" db:"id"`
	Token          string `json:"jwt" db:"token"`
	UserId         int    `json:"user_id" db:"user_id"`
	ExpireAt       int64  `json:"-" db:"expire_at"`
	ExpireDuration int    `json:"expire_duration"`
}

type AccessToken struct {
	Username       string `json:"username"`
	RefreshTokenId int    `json:"refresh_token_id"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
