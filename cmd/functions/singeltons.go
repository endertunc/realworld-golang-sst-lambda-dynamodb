package functions

import "realworld-aws-lambda-dynamodb-golang/internal/user"

var (
	UserRepository     = user.DynamodbUserRepository{}
	UserService        = user.UserService{UserRepository: UserRepository, FollowerService: FollowerService}
	FollowerRepository = user.DynamodbFollowerRepository{}
	FollowerService    = user.FollowerService{FollowerRepository: FollowerRepository, UserService: UserService}
	FollowerApi        = user.FollowerApi{FollowerService: FollowerService}
	UserApi            = user.UserApi{UserService: UserService}
)