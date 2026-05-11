package auth

import "gofiber-starterkit/app/models"

func TransformUser(user *models.User) LoginResponse {
	return LoginResponse{
		User_ID:  user.ID.String(),
		Username: user.Username,
	}
}
