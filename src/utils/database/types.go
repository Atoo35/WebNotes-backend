package database

type User struct {
	Email        string `json:"email" db:"email"`
	AccessToken  string `json:"access_token" db:"access_token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
}
