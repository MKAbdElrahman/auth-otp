package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "midaslabs/gen/auth/v1"

	"midaslabs/microservices/auth/internal/application"
	"midaslabs/microservices/auth/internal/domain"
)

type AuthServerHandlers struct {
	authService *application.AuthService
	logger      *log.Logger
}

func NewAuthServerHandlers(logger *log.Logger, authService *application.AuthService) *AuthServerHandlers {
	return &AuthServerHandlers{
		authService: authService,
		logger:      logger,
	}
}

func (s *AuthServerHandlers) SignUpWithPhoneNumber(
	ctx context.Context,
	req *connect.Request[authv1.SignUpWithPhoneNumberRequest],
) (*connect.Response[authv1.SignUpWithPhoneNumberResponse], error) {
	err := s.authService.SignUpWithPhoneNumber(ctx, req.Msg.Phone)
	if err != nil {
		s.logger.Errorf("SignUpWithPhoneNumber: failed to sign up for phone number %s: %v", req.Msg.Phone, err)
		if err == domain.ErrUserAlreadyExists {
			return connect.NewResponse(&authv1.SignUpWithPhoneNumberResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "User already exists",
					ErrorCode: "ERR_USER_EXISTS",
				},
			}), nil

		}
		return connect.NewResponse(&authv1.SignUpWithPhoneNumberResponse{
			Status: &authv1.ResponseStatus{
				Success:   false,
				Message:   "Failed to sign up",
				ErrorCode: "ERR_INTERNAL",
			},
		}), nil
	}

	s.logger.Infof("SignUpWithPhoneNumber: OTP sent to phone number %s", req.Msg.Phone)
	return connect.NewResponse(&authv1.SignUpWithPhoneNumberResponse{
		Status: &authv1.ResponseStatus{
			Success: true,
			Message: "OTP sent",
		},
	}), nil
}

func (s *AuthServerHandlers) VerifyPhoneNumber(
	ctx context.Context,
	req *connect.Request[authv1.VerifyPhoneNumberRequest],
) (*connect.Response[authv1.VerifyPhoneNumberResponse], error) {
	err := s.authService.VerifyPhoneNumber(ctx, req.Msg.Phone, req.Msg.Otp)
	if err != nil {
		s.logger.Errorf("VerifyPhoneNumber: failed to verify phone number %s: %v", req.Msg.Phone, err)
		if err == domain.ErrInvalidOTP {
			return connect.NewResponse(&authv1.VerifyPhoneNumberResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "Invalid OTP",
					ErrorCode: "ERR_INVALID_OTP",
				},
			}), nil
		}
		return connect.NewResponse(&authv1.VerifyPhoneNumberResponse{
			Status: &authv1.ResponseStatus{
				Success:   false,
				Message:   "Failed to verify phone number",
				ErrorCode: "ERR_INTERNAL",
			},
		}), nil
	}

	s.logger.Infof("VerifyPhoneNumber: phone number %s verified", req.Msg.Phone)
	return connect.NewResponse(&authv1.VerifyPhoneNumberResponse{
		Status: &authv1.ResponseStatus{
			Success: true,
			Message: "Phone number verified",
		},
	}), nil
}

func (s *AuthServerHandlers) LoginInitiate(
	ctx context.Context,
	req *connect.Request[authv1.LoginInitiateRequest],
) (*connect.Response[authv1.LoginInitiateResponse], error) {
	err := s.authService.LoginInitiate(ctx, req.Msg.Phone)
	if err != nil {
		s.logger.Errorf("LoginInitiate: failed to initiate login for phone number %s: %v", req.Msg.Phone, err)
		if err == domain.ErrUserNotFound {
			return connect.NewResponse(&authv1.LoginInitiateResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "User not found",
					ErrorCode: "ERR_USER_NOT_FOUND",
				},
			}), nil
		}
		return connect.NewResponse(&authv1.LoginInitiateResponse{
			Status: &authv1.ResponseStatus{
				Success:   false,
				Message:   "Failed to initiate login",
				ErrorCode: "ERR_INTERNAL",
			},
		}), nil
	}

	s.logger.Infof("LoginInitiate: OTP sent to phone number %s", req.Msg.Phone)
	return connect.NewResponse(&authv1.LoginInitiateResponse{
		Status: &authv1.ResponseStatus{
			Success: true,
			Message: "OTP sent",
		},
	}), nil
}

func (s *AuthServerHandlers) ValidatePhoneNumberLogin(
	ctx context.Context,
	req *connect.Request[authv1.ValidatePhoneNumberLoginRequest],
) (*connect.Response[authv1.ValidatePhoneNumberLoginResponse], error) {
	err := s.authService.ValidatePhoneNumberLogin(ctx, req.Msg.Phone, req.Msg.Otp)
	if err != nil {
		s.logger.Errorf("ValidatePhoneNumberLogin: failed to login with phone number %s: %v", req.Msg.Phone, err)
		if err == domain.ErrInvalidOTP {
			return connect.NewResponse(&authv1.ValidatePhoneNumberLoginResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "Invalid OTP",
					ErrorCode: "ERR_INVALID_OTP",
				},
			}), nil
		} else if err == domain.ErrUserNotVerified {
			return connect.NewResponse(&authv1.ValidatePhoneNumberLoginResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "User not verified",
					ErrorCode: "ERR_USER_NOT_VERIFIED",
				},
			}), nil
		} else if err == domain.ErrUserNotFound {
			return connect.NewResponse(&authv1.ValidatePhoneNumberLoginResponse{
				Status: &authv1.ResponseStatus{
					Success:   false,
					Message:   "User not found",
					ErrorCode: "ERR_USER_NOT_FOUND",
				},
			}), nil
		}
		return connect.NewResponse(&authv1.ValidatePhoneNumberLoginResponse{
			Status: &authv1.ResponseStatus{
				Success:   false,
				Message:   "Failed to login",
				ErrorCode: "ERR_INTERNAL",
			},
		}), nil
	}

	s.logger.Infof("ValidatePhoneNumberLogin: user %s logged in successfully", req.Msg.Phone)
	return connect.NewResponse(&authv1.ValidatePhoneNumberLoginResponse{
		Status: &authv1.ResponseStatus{
			Success: true,
			Message: "Login successful",
		},
	}), nil
}
func (s *AuthServerHandlers) GetProfile(
	ctx context.Context,
	req *connect.Request[authv1.GetProfileRequest],
) (*connect.Response[authv1.GetProfileResponse], error) {
	user, err := s.authService.GetProfile(ctx, req.Msg.Phone)
	if err != nil {
		s.logger.Errorf("GetProfile: failed to get profile for phone number %s: %v", req.Msg.Phone, err)
		return connect.NewResponse(&authv1.GetProfileResponse{
			Status: &authv1.ResponseStatus{
				Success:   false,
				Message:   "Failed to get profile",
				ErrorCode: "ERR_INTERNAL",
			},
		}), nil
	}

	profileData := &authv1.ProfileData{
		PhoneNumber: user.PhoneNumber,
		Verified:    user.Verified,
		CreatedAt:   &timestamppb.Timestamp{Seconds: user.CreatedAt.Unix()},
		UpdatedAt:   &timestamppb.Timestamp{Seconds: user.UpdatedAt.Unix()},
	}

	s.logger.Infof("GetProfile: retrieved profile for phone number %s", req.Msg.Phone)
	return connect.NewResponse(&authv1.GetProfileResponse{
		Status: &authv1.ResponseStatus{
			Success: true,
			Message: "Profile retrieved successfully",
		},
		ProfileData: profileData,
	}), nil
}
