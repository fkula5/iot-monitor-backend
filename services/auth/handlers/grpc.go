package handlers

import (
	"context"
	"log"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	"github.com/skni-kod/iot-monitor-backend/services/auth/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthGrpcHandler struct {
	pb.UnimplementedAuthServiceServer
	authService services.IAuthService
}

func NewGrpcHandler(s *grpc.Server, authService services.IAuthService) {
	handler := &AuthGrpcHandler{
		authService: authService,
	}
	pb.RegisterAuthServiceServer(s, handler)
}

func (h *AuthGrpcHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters long")
	}

	authReq := &services.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	authRes, err := h.authService.Register(ctx, authReq)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Token:     authRes.Token,
		ExpiresAt: timestamppb.New(authRes.ExpiresAt),
		User:      convertUserToProto(&authRes.User),
	}, nil
}

func convertUserToProto(userInfo *services.UserInfo) *pb.User {
	if userInfo == nil {
		return nil
	}

	return &pb.User{
		Id:        int32(userInfo.ID),
		Email:     userInfo.Email,
		Username:  userInfo.Username,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Active:    true,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}
}

func convertEntUserToProto(user *ent.User) *pb.User {
	if user == nil {
		return nil
	}

	return &pb.User{
		Id:        int32(user.ID),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
