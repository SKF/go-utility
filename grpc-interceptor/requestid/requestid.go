package requestid

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/SKF/go-utility/uuid"
)

const REQUEST_ID_KEY = "request.id"
const REQUEST_CHAIN_KEY = "request.chain"
const REQUEST_TRANSACTION_ID_KEY = "request.transaction.id"

// UnaryServerInterceptor returns a new unary server interceptor that adds
// the Request ID Metadata to the call.
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		newCtx := ExtendContext(ctx, serviceName)
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds
// the Request ID Metadata to the call.
func StreamServerInterceptor(serviceName string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		newCtx := ExtendContext(stream.Context(), serviceName)
		wrappedStream := grpc_middleware.WrapServerStream(stream)
		wrappedStream.WrappedContext = newCtx
		return handler(srv, wrappedStream)
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that adds
// the Request ID Metadata to the call.
func UnaryClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx := outgoingContextWithRequestID(ctx, serviceName)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that adds
// the Request ID Metadata to the call.
func StreamClientInterceptor(serviceName string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newCtx := outgoingContextWithRequestID(ctx, serviceName)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

// ExtendContext extends the context with a Request ID Metadata.
func ExtendContext(ctx context.Context, serviceName string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	transactionID := uuid.New()
	request := Request{
		ID:            transactionID,
		Chain:         []string{serviceName},
		TransactionID: transactionID,
	}

	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ids := incomingMD.Get(REQUEST_ID_KEY)
		if len(ids) > 0 {
			id := uuid.UUID(ids[0])
			if id.IsValid() {
				request.ID = id
			}
		}

		request.Chain = append(incomingMD.Get(REQUEST_CHAIN_KEY), serviceName)
	}

	outgoingMD := make(metadata.MD)
	outgoingMD.Set(REQUEST_ID_KEY, request.ID.String())
	outgoingMD.Set(REQUEST_CHAIN_KEY, request.Chain...)
	outgoingMD.Set(REQUEST_TRANSACTION_ID_KEY, request.TransactionID.String())
	return metadata.NewOutgoingContext(ctx, outgoingMD)
}

func outgoingContextWithRequestID(ctx context.Context, serviceName string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	transactionID := uuid.New()
	request := Request{
		ID:            transactionID,
		Chain:         []string{serviceName},
		TransactionID: transactionID,
	}

	incomingMD, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		ids := incomingMD.Get(REQUEST_ID_KEY)
		if len(ids) > 0 {
			id := uuid.UUID(ids[0])
			if id.IsValid() {
				request.ID = id
			}
		}
		ids = incomingMD.Get(REQUEST_TRANSACTION_ID_KEY)
		if len(ids) > 0 {
			id := uuid.UUID(ids[0])
			if id.IsValid() {
				request.TransactionID = id
			}
		}

		request.Chain = append(incomingMD.Get(REQUEST_CHAIN_KEY))
	}

	outgoingMD := make(metadata.MD)
	outgoingMD.Set(REQUEST_ID_KEY, request.ID.String())
	outgoingMD.Set(REQUEST_CHAIN_KEY, request.Chain...)
	outgoingMD.Set(REQUEST_TRANSACTION_ID_KEY, request.TransactionID.String())
	return metadata.NewOutgoingContext(ctx, outgoingMD)
}

// Request is a data holder for the different Request ID Metadata
type Request struct {
	ID            uuid.UUID `json:"id"`
	Chain         []string  `json:"chain"`
	TransactionID uuid.UUID `json:"transactionId"`
}

// Extract will get the Request ID Metadata out of the context.
func Extract(ctx context.Context) (request Request) {
	if ctx == nil {
		return
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return
	}

	ids := md.Get(REQUEST_ID_KEY)
	if len(ids) > 0 {
		id := uuid.UUID(ids[0])
		if id.IsValid() {
			request.ID = id
		}
	}

	ids = md.Get(REQUEST_TRANSACTION_ID_KEY)
	if len(ids) > 0 {
		id := uuid.UUID(ids[0])
		if id.IsValid() {
			request.TransactionID = id
		}
	}

	request.Chain = append(md.Get(REQUEST_CHAIN_KEY))
	return
}
