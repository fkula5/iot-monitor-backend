package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptors that logs incoming requests.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		st, _ := status.FromError(err)

		L().Info("gRPC request",
			zap.String("grpc.method", info.FullMethod),
			zap.String("grpc.code", st.Code().String()),
			zap.Duration("grpc.duration", duration),
			zap.Error(err),
		)

		return resp, err
	}
}
