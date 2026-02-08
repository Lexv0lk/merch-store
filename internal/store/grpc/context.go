package grpc

var userIdContextKey = contextKey{name: "user_id"}
var usernameContextKey = contextKey{name: "username"}

type contextKey struct {
	name string
}
