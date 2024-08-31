# reference file: https://github.com/aws-samples/serverless-go-demo/blob/main/Makefile

FUNCTIONS := add_comment create_article delete_article delete_comment favorite_article follow_user get_article get_article_comments get_current_user get_user_profile login_user register_user unfavorite_article unfollow_user update_article update_user

# To try different version of Go
GO := go

# Make sure to install aarch64 GCC compilers if you want to compile with GCC.
CC := aarch64-linux-gnu-gcc
GCCGO := aarch64-linux-gnu-gccgo-10

build:
		${MAKE} ${MAKEOPTS} $(foreach function,${FUNCTIONS}, build-${function})

build-%:
		cd cmd/functions/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-gcc:
		${MAKE} ${MAKEOPTS} $(foreach function,${FUNCTIONS}, build-gcc-${function})

build-gcc-%:
		cd functions/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=${CC} ${GO} build -o bootstrap

build-gcc-optimized:
		${MAKE} ${MAKEOPTS} $(foreach function,${FUNCTIONS}, build-gcc-optimized-${function})

build-gcc-optimized-%:
		cd functions/$* && GOOS=linux GOARCH=arm64 GCCGO=${GCCGO} ${GO} build -compiler gccgo -gccgoflags '-static -Ofast -march=armv8.2-a+fp16+rcpc+dotprod+crypto -mtune=neoverse-n1 -moutline-atomics' -o bootstrap

clean:
	@rm $(foreach function,${FUNCTIONS}, cmd/functions/${function}/bootstrap)

tests-unit:
	@go test -v -tags=unit -bench=. -benchmem -cover ./...



.PHONY: tidy
tidy:
	@$(foreach dir,$(MODULE_DIRS),(cd $(dir) && go mod tidy) &&) true




# List of main functions to build
#MAIN_FUNCTIONS := \
#	add_comment \
#	create_article \
#	create_article \
#	delete_article \
#	delete_comment \
#	favorite_article \
#	follow_user \
#	get_article \
#	get_article_comments \
#	get_current_user \
#	get_user_profile \
#	login_user \
#	register_user \
#	unfavorite_article \
#	unfollow_user \
#	update_article \
#	update_user \