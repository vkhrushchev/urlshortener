package interceptor

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/vkhrushchev/urlshortener/internal/common"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"slices"
	"strings"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

func UserIDInterceptor(salt string, acceptedMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		method := info.FullMethod[strings.LastIndexByte(info.FullMethod, '/')+1:]
		if slices.Contains(acceptedMethods, method) {
			log.Infow("interceptor: calling UserIDInterceptor", "method", method)

			var userID, userIDSignature string
			var isValidSignature bool

			userIDMetadata := metadata.ValueFromIncomingContext(ctx, "user-id")
			userIDSignatureMetadata := metadata.ValueFromIncomingContext(ctx, "user-id-signature")

			if len(userIDMetadata) == 1 && len(userIDSignatureMetadata) == 1 {
				userID = userIDMetadata[0]
				userIDSignature = userIDSignatureMetadata[0]
				isValidSignature = common.CheckSignature(userID, userIDSignature, salt)
			}

			if len(userIDMetadata) != 1 || len(userIDSignatureMetadata) != 1 || !isValidSignature {
				log.Infow("interceptor: 'user-id' metadata not found or not valid")

				userID = uuid.NewString()
				userIDSignatureBytes := md5.Sum([]byte(userID + salt))
				userIDSignature = hex.EncodeToString(userIDSignatureBytes[:])

			}

			ctx = context.WithValue(ctx, common.UserIDContextKey, userID)
			res, err := handler(ctx, req)

			grpc.SetHeader(ctx, metadata.Pairs("user-id", userID, "user-id-signature", userIDSignature))

			return res, err
		}

		return handler(ctx, req)
	}
}

func AuthByUserIDInterceptor(salt string, acceptedMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		method := info.FullMethod[strings.LastIndexByte(info.FullMethod, '/')+1:]
		if slices.Contains(acceptedMethods, method) {
			log.Infow("interceptor: calling AuthByUserIDInterceptor", "method", method)

			var userID, userIDSignature string
			var isValidSignature bool

			userIDMetadata := metadata.ValueFromIncomingContext(ctx, "user-id")
			userIDSignatureMetadata := metadata.ValueFromIncomingContext(ctx, "user-id-signature")

			if len(userIDMetadata) == 1 && len(userIDSignatureMetadata) == 1 {
				userID = userIDMetadata[0]
				userIDSignature = userIDSignatureMetadata[0]
				isValidSignature = common.CheckSignature(userID, userIDSignature, salt)
			}

			if !isValidSignature {
				log.Infow("interceptor: 'user-id-signature' metadata not found or not valid")
				return nil, status.Errorf(codes.Unauthenticated, "invalid 'user-id-signature' metadata")
			}
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			grpc.SetHeader(ctx, md)
		}

		return handler(ctx, req)
	}
}

func CheckSubnetInterceptor(trustedSubnet *net.IPNet, acceptedMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		method := info.FullMethod[strings.LastIndexByte(info.FullMethod, '/')+1:]
		if slices.Contains(acceptedMethods, method) {
			log.Infow("interceptor: calling CheckSubnetInterceptor", "method", method)

			var xRealIP string

			xRealIPMetadata := metadata.ValueFromIncomingContext(ctx, "x-real-ip")
			if len(xRealIPMetadata) == 1 {
				xRealIP = xRealIPMetadata[0]
			}

			if trustedSubnet == nil || !trustedSubnet.Contains(net.ParseIP(xRealIP)) {
				return nil, status.Errorf(codes.Unauthenticated, "invalid 'x-real-ip' metadata")
			}
		}

		return handler(ctx, req)
	}
}
