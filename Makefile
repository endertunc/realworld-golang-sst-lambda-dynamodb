# reference file: https://github.com/aws-samples/serverless-go-demo/blob/main/Makefile

GO := go
FUNCTIONS := add_comment delete_article delete_comment favorite_article follow_user get_article get_article_comments get_current_user get_user_feed get_user_profile list_articles login_user post_article register_user unfavorite_article unfollow_user update_article update_user user_feed

build:
		${MAKE} ${MAKEOPTS} $(foreach function,${FUNCTIONS}, build-${function})

build-%:
		cd cmd/functions/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

test-e2e:
		${GO} test ./cmd/functions/... -p 1 -v

test-unit:
		${GO} test ./internal/service/ -p 1 -v -cover

clean:
	@rm $(foreach function,${FUNCTIONS}, cmd/functions/${function}/bootstrap)