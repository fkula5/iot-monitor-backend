package handlers

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	"github.com/skni-kod/iot-monitor-backend/services/auth/services"
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
	logger.Info("Auth gRPC handler registered")
}

func (h *AuthGrpcHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" {
		logger.Warn("Login request missing email")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		logger.Warn("Login request missing password")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	authReq := &services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	authRes, err := h.authService.Login(ctx, authReq)
	if err != nil {
		logger.Error("Failed to login user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{
		Token:     authRes.Token,
		ExpiresAt: timestamppb.New(authRes.ExpiresAt),
		User:      convertUserToProto(&authRes.User),
	}, nil
}

func (h *AuthGrpcHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" {
		logger.Warn("Register request missing email")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Username == "" {
		logger.Warn("Register request missing username")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Password == "" {
		logger.Warn("Register request missing password")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.Password) < 8 {
		logger.Warn("Register request password too short")
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
		logger.Error("Failed to register user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Token:     authRes.Token,
		ExpiresAt: timestamppb.New(authRes.ExpiresAt),
		User:      convertUserToProto(&authRes.User),
	}, nil
}

func (h *AuthGrpcHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := h.authService.GetUserByID(ctx, int(req.Id))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserResponse{
		User: convertEntUserToProto(user),
	}, nil
}

func (h *AuthGrpcHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	updateData := &services.UpdateRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := h.authService.Update(ctx, int(req.Id), updateData)

	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UserResponse{
		User: convertEntUserToProto(user),
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
