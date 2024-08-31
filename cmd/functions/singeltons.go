package functions

import "realworld-aws-lambda-dynamodb-golang/internal/user"

var (
	UserRepository     = user.DynamodbUserRepository{}
	UserService        = user.UserService{UserRepository: UserRepository}
	FollowerRepository = user.DynamodbFollowerRepository{}
	FollowerService    = user.FollowerService{FollowerRepository: FollowerRepository, UserService: UserService}
	FollowerApi        = user.FollowerApi{FollowerService: FollowerService}
	UserApi            = user.UserApi{UserService: UserService}
	ArticleApi         = user.ArticleApi{UserService: UserService}
)
