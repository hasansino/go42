// #nosec

package integration

import (
	"fmt"
	"math/rand"
	"os"
)

const (
	httpServerAddressEnvVarName = "HTTP_SERVER_ADDRESS"
	grpcServerAddressEnvVarName = "GRPC_SERVER_ADDRESS"
)

const (
	defaultHttpServerAddress = "http://localhost:8080/api/v1"
	defaultGrpcServerAddress = "localhost:50051"
)

var (
	customHttpServerAddress string
	customGrpcServerAddress string
)

func init() {
	value, found := os.LookupEnv(httpServerAddressEnvVarName)
	if found {
		customHttpServerAddress = value
	}
	value, found = os.LookupEnv(grpcServerAddressEnvVarName)
	if found {
		customGrpcServerAddress = value
	}
}

func HTTPServerAddress() string {
	if customHttpServerAddress != "" {
		return customHttpServerAddress
	}
	return defaultHttpServerAddress
}

func GRPCServerAddress() string {
	if customGrpcServerAddress != "" {
		return customGrpcServerAddress
	}
	return defaultGrpcServerAddress
}

// ---

const randomStringDefaultLength = 8

// GenerateRandomString returns a unique fruit name with given prefix.
func GenerateRandomString(prefix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, randomStringDefaultLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}
