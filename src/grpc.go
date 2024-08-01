package nibisdk

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GetGRPCConnection establishes a connection to a gRPC server using either
// secure (TLS) or insecure credentials. The function blocks until the connection
// is established or the specified timeout is reached.
func GetGRPCConnection(
	grpcUrl string, grpcInsecure bool, timeout time.Duration,
) *grpc.ClientConn {
	var creds credentials.TransportCredentials
	if grpcInsecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{})
	}

	options := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(creds),
	}
	ctx, cancel := context.WithTimeout(
		context.Background(), timeout,
	)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcUrl, options...)
	if err != nil {
		panic(err)
	}

	return conn
}
