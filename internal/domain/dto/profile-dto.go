package dto

import "realworld-aws-lambda-dynamodb-golang/internal/domain"

type ProfileResponseBodyDTO struct {
	Profile ProfileResponseDto `json:"profile"`
}

type ProfileResponseDto struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

func ToProfileResponseBodyDTO(user domain.User, isFollowing bool) ProfileResponseBodyDTO {
	return ProfileResponseBodyDTO{
		Profile: ProfileResponseDto{
			Username:  user.Username,
			Bio:       user.Bio,
			Image:     user.Image,
			Following: isFollowing,
		},
	}
}
