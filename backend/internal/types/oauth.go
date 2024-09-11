package types

// OAuthUserInfo represents the structure of user info from an OAuth provider
type OAuthUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
