package dto

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

// login user request dtos
type LoginRequestBodyDTO struct {
	User LoginRequestUserDto `json:"user" validate:"required"`
}

type LoginRequestUserDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,notblank,min=6,max=20"`
}

func (s LoginRequestBodyDTO) Validate() ValidationErrors {
	return validateStruct(s)
}

// new user request dtos
type NewUserRequestBodyDTO struct {
	User NewUserRequestUserDto `json:"user" validate:"required"`
}

type NewUserRequestUserDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,notblank,min=6,max=20"`
	Username string `json:"username" validate:"required,notblank,min=3,max=64"`
}

func (s NewUserRequestBodyDTO) Validate() ValidationErrors {
	return validateStruct(s)
}

// update user request dtos
type UpdateUserRequestBodyDTO struct {
	User UpdateUserRequestUserDTO `json:"user" validate:"required"`
}

type UpdateUserRequestUserDTO struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	Password *string `json:"password" validate:"omitempty,notblank,min=6,max=20"`
	Username *string `json:"username" validate:"omitempty,notblank,min=3,max=64"`
	Bio      *string `json:"bio" validate:"omitempty"`
	Image    *string `json:"image" validate:"omitempty,url"`
}

func (s UpdateUserRequestBodyDTO) Validate() ValidationErrors {
	return validateStruct(s)
}

// user response dtos
type UserResponseBodyDTO struct {
	User UserResponseUserDto `json:"user"`
}
type UserResponseUserDto struct {
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Token    string  `json:"token"`
	Bio      *string `json:"bio"`
	// ToDo @ender - once we have update profile, we should validate that this is a valid url I guess?
	//   It's not clear to me what this field suppose to store. I am assuming it's just a url to the image.
	Image *string `json:"image"`
}

func ToUserResponseBodyDTO(user domain.User, token domain.Token) UserResponseBodyDTO {
	return UserResponseBodyDTO{
		User: UserResponseUserDto{
			Email:    user.Email,
			Username: user.Username,
			Token:    string(token),
			Bio:      user.Bio,
			Image:    user.Image,
		},
	}
}
