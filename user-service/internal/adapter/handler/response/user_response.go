package response

type SignInResponse struct {
	AccessToken string `json:"access_token"`
	Role        string `json:"role"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
}

type ProfileResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	RoleName string `json:"role"`
	Phone    string `json:"phone"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Address  string `json:"address"`
	Photo    string `json:"photo"`
}

type CustomerListResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Photo string `json:"photo"`
}

type CustomerResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Address  string `json:"address"`
	Photo    string `json:"photo"`
	RoleID   int64  `json:"role_id"`
	RoleName string `json:"role,omitempty"`
}
