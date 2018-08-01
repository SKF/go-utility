package requestid_test

import (
	"context"

	"github.com/SKF/go-utility/grpc_interceptor/requestid"
	"github.com/SKF/go-utility/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
)

func Example() {
	_ = grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			requestid.StreamServerInterceptor("LOG_NAME"),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			requestid.UnaryServerInterceptor("LOG_NAME"),
		)),
	)
}

func ExampleExtract() {
	var grpcCallContext context.Context
	log.WithField("request", requestid.Extract(grpcCallContext)).
		Infof("Request ID Metadata")
}

func ExampleRequest_NewOutgoingContext() {
	var grpcCallContext context.Context

	outgoingGrpcCallContext := context.Background()
	outgoingGrpcCallContext = requestid.
		Extract(grpcCallContext).
		NewOutgoingContext(outgoingGrpcCallContext)
}
