package interceptors

var DefaultSkipper = func(method string) bool {
	if method == "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo" ||
		method == "/grpc.health.v1.Health/List" ||
		method == "/grpc.health.v1.Health/Check" ||
		method == "/grpc.health.v1.Health/Watch" {
		return true
	}
	return false
}
