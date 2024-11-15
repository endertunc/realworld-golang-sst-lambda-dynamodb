package generator

import (
	"github.com/brianvoe/gofakeit/v7"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

func GenerateNewUserRequestBodyDTO() dto.NewUserRequestBodyDTO {
	return dto.NewUserRequestBodyDTO{
		User: GenerateNewUserRequestUserDto(),
	}
}

func GenerateNewUserRequestUserDto() dto.NewUserRequestUserDto {
	return dto.NewUserRequestUserDto{
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, false, gofakeit.IntRange(6, 20)),
		Username: gofakeit.Username(),
	}
}
