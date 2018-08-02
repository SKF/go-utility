package requestid

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
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
		addRequestID(ctx, serviceName)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds
// the Request ID Metadata to the call.
func StreamServerInterceptor(serviceName string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		addRequestID(stream.Context(), serviceName)
		return handler(srv, stream)
	}
}

func addRequestID(ctx context.Context, serviceName string) {
	tags := grpc_ctxtags.Extract(ctx)
	req := New(serviceName)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ids := md.Get(REQUEST_ID_KEY)
		if len(ids) > 0 {
			id := uuid.UUID(ids[0])
			if id.IsValid() {
				req.ID = id
			}
		}

		req.Chain = append(md.Get(REQUEST_CHAIN_KEY), serviceName)
	}

	tags.Set(REQUEST_ID_KEY, req.ID)
	tags.Set(REQUEST_CHAIN_KEY, req.Chain)
	tags.Set(REQUEST_TRANSACTION_ID_KEY, req.TransactionID)
	return
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

	tags := grpc_ctxtags.Extract(ctx)

	if value, exists := tags.Values()[REQUEST_ID_KEY]; exists {
		request.ID = value.(uuid.UUID)
	}

	if value, exists := tags.Values()[REQUEST_CHAIN_KEY]; exists {
		request.Chain = value.([]string)
	}

	if value, exists := tags.Values()[REQUEST_TRANSACTION_ID_KEY]; exists {
		request.TransactionID = value.(uuid.UUID)
	}

	return
}

func New(serviceName string) Request {
	transactionID := uuid.New()
	return Request{
		ID:            transactionID,
		Chain:         []string{serviceName},
		TransactionID: transactionID,
	}
}

// NewOutgoingContext creates a new context with the outgoing
// Request ID Metadata attached.
func (request Request) NewOutgoingContext(ctx context.Context) context.Context {
	md := make(metadata.MD)
	md.Set(REQUEST_ID_KEY, request.ID.String())
	md.Set(REQUEST_CHAIN_KEY, request.Chain...)
	md.Set(REQUEST_TRANSACTION_ID_KEY, request.TransactionID.String())
	return metadata.NewOutgoingContext(ctx, md)
}
