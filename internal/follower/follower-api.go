package follower

import (
	"net/http"
)

type FollowerApi struct {
}

func (fa FollowerApi) UnfollowUserByUsername(w http.ResponseWriter, r *http.Request, username string) {
	//TODO implement me
	panic("implement me")
}

func (fa FollowerApi) FollowUserByUsername(w http.ResponseWriter, r *http.Request, username string) {
	//TODO implement me
	panic("implement me")
}
