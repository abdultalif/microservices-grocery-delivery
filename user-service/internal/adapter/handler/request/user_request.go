package request

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"password,omitempty"`
	NewPassword     string `json:"password_new" validate:"required"`
	ConfirmPassword string `json:"password_confirmation" validate:"required"`
}

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required,min=8"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

type UpdateDataUserRequest struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Phone   int64   `json:"phone"`
	Address string  `json:"address"`
}