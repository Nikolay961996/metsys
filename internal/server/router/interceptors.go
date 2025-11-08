// Package router consist interceptors
package router

import (
	"context"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func WithTrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
		xRealIP := getClientIPFromContextGRPC(ctx)
		code := checkTrustedSubnet(xRealIP, trustedSubnet)

		switch code {
		case codes.PermissionDenied:
			return nil, status.Error(code, "Forbidden: IP not in trusted subnet")
		case codes.Internal:
			return nil, status.Error(code, "Server configuration error")
		case codes.OK:
			return next(ctx, req)
		default:
			models.Log.Warn(fmt.Sprintf("Not expectes code: %d", code))
			return next(ctx, req)
		}
	}
}

func getClientIPFromContextGRPC(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("X-Real-IP")
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
