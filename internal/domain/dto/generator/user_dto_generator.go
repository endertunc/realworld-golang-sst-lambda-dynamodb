package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

func GenerateNewUserRequestUserDto() dto.NewUserRequestUserDto {
	return dto.NewUserRequestUserDto{
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, false, gofakeit.IntRange(6, 20)),
		Username: gofakeit.Username(),
	}
}

func GenerateLoginRequestBodyDto() dto.LoginRequestUserDto {
	return dto.LoginRequestUserDto{
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, false, gofakeit.IntRange(6, 20)),
	}
}

func GenerateUpdateUserRequestUserDTO() dto.UpdateUserRequestUserDTO {
	email := gofakeit.Email()
	username := gofakeit.Username()
	password := gofakeit.Password(true, true, true, true, false, gofakeit.IntRange(6, 20))
	bio := gofakeit.Sentence(10)
	image := gofakeit.URL()

	return dto.UpdateUserRequestUserDTO{
		Email:    &email,
		Username: &username,
		Password: &password,
		Bio:      &bio,
		Image:    &image,
	}
}
