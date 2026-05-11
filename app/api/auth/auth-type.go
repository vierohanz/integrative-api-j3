package auth

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User_ID  string `json:"user_id"`
	Username string `json:"username"`
}
