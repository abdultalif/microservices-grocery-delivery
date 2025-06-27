package entity

type UserEntity struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	RoleName   string `json:"role_name"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	IsVerified bool   `json:"is_verified"`
	Photo      string `json:"photo"`
	Lat        string `json:"lat"`
	Lng        string `json:"lng"`
	Token      string
}