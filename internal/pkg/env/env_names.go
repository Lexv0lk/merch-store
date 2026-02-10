package env

const (
	EnvGrpcAuthPort  = "GRPC_AUTH_PORT"
	EnvGrpcStorePort = "GRPC_STORE_PORT"
	EnvHttpPort      = "HTTP_PORT"

	EnvAuthDatabaseHost     = "DB_AUTH_HOST"
	EnvAuthDatabasePort     = "DB_AUTH_PORT"
	EnvAuthDatabaseUser     = "DB_AUTH_USER"
	EnvAuthDatabasePassword = "DB_AUTH_PASSWORD"
	EnvAuthDatabaseName     = "DB_AUTH_NAME"

	EnvStoreDatabaseHost     = "DB_STORE_HOST"
	EnvStoreDatabasePort     = "DB_STORE_PORT"
	EnvStoreDatabaseUser     = "DB_STORE_USER"
	EnvStoreDatabasePassword = "DB_STORE_PASSWORD"
	EnvStoreDatabaseName     = "DB_STORE_NAME"

	EnvJwtSecret = "JWT_SECRET"

	EnvGrpcAuthHost  = "GRPC_AUTH_HOST"
	EnvGrpcStoreHost = "GRPC_STORE_HOST"
)
