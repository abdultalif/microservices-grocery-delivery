package entity

type JwtUserData struct {
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	LoggedIn  bool   `json:"logged_in"`
	CreatedAt string `json:"created_at"`
	Token     string `json:"token"`
}
