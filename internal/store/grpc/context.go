package grpc

import "time"

const contextTimeLimit = 500 * time.Millisecond

var userIdContextKey = contextKey{name: "user_id"}

type contextKey struct {
	name string
}
