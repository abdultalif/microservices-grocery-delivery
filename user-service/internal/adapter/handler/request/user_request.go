package request

type TokenRequest struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdatePasswordRequest struct {
	NewPassword     string `json:"password_new" validate:"required,min=8"`
	ConfirmPassword string `json:"password_confirmation" validate:"required,min=8"`
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

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"password_new" validate:"required,min=8"`
	ConfirmPassword string `json:"password_confirmation" validate:"required,min=8"`
}