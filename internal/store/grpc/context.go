package grpc

import "time"

const contextTimeLimit = 500 * time.Millisecond

var userIdContextKey = contextKey{name: "user_id"}
var usernameContextKey = contextKey{name: "username"}

type contextKey struct {
	name string
}
