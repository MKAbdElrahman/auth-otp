package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"midaslabs/microservices/auth/internal/domain"
	"time"
)

const otpValidityDuration = 5 * time.Minute

// AuthService handles user authentication and OTP operations.
type AuthService struct {
	userRepo      domain.UserRepository
	activityRepo  domain.ActivityRepository
	otpRepo       domain.OTPRepository
	messageBroker domain.MessageBroker
}

func NewAuthService(userRepo domain.UserRepository, activityRepo domain.ActivityRepository, otpRepo domain.OTPRepository, messageBroker domain.MessageBroker) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		activityRepo:  activityRepo,
		otpRepo:       otpRepo,
		messageBroker: messageBroker,
	}
}

// SignUpWithPhoneNumber handles user signup and requests OTP for verification.
func (s *AuthService) SignUpWithPhoneNumber(ctx context.Context, phoneNumber string) error {
	_, err := s.userRepo.GetUser(ctx, phoneNumber)
	if err == nil {
		return domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}

	user := domain.NewUser(phoneNumber)
	if err := s.userRepo.AddUser(ctx, user); err != nil {
		return err
	}

	// Request OTP for verification
	if err := s.requestNewOTP(ctx, phoneNumber); err != nil {
		return err
	}

	// Log the signup activity
	activity := &domain.Activity{
		PhoneNumber: phoneNumber,
		Type:        domain.ActivitySignup,
		Timestamp:   time.Now(),
	}
	if err := s.activityRepo.RecordActivity(ctx, activity); err != nil {
		return err
	}

	return nil
}

// requestOTP generates a new OTP, deletes any existing one, and sends it via the message broker.
func (s *AuthService) requestNewOTP(ctx context.Context, phoneNumber string) error {
	// Delete old OTP if exists
	if err := s.otpRepo.DeleteOTP(ctx, phoneNumber); err != nil && !errors.Is(err, domain.ErrOTPNotFound) {
		return err
	}

	// Generate new OTP
	otpCode, err := generateOTP()
	if err != nil {
		return err
	}

	expiration := time.Now().Add(otpValidityDuration)

	// Save new OTP to the database
	if err := s.otpRepo.StoreOTP(ctx, phoneNumber, otpCode, expiration); err != nil {
		return err
	}

	// Publish OTP message
	if err := s.publishSendOTPEvent(ctx, phoneNumber, otpCode); err != nil {
		return err
	}

	return nil
}

// publishSendOTPEvent sends an OTP message to the message broker.
func (s *AuthService) publishSendOTPEvent(ctx context.Context, phoneNumber, otpCode string) error {
	// Create a message payload
	event := domain.OTPVerificationEvent{
		PhoneNumber: phoneNumber,
		OTPCode:     otpCode,
	}
	message, err := event.Serialize()
	if err != nil {
		return err
	}

	// Publish message to the broker
	if err := s.messageBroker.Publish(ctx, "verification", message); err != nil {
		return err
	}
	return nil
}

func generateOTP() (string, error) {
	b := make([]byte, 4) // 4 bytes to generate an OTP of 8 hexadecimal characters
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// VerifyPhoneNumber verifies the OTP for the given phone number after signup.
func (s *AuthService) VerifyPhoneNumber(ctx context.Context, phoneNumber, otp string) error {
	// Retrieve OTP from the database
	storedOTP, expiration, err := s.otpRepo.GetOTP(ctx, phoneNumber)
	if err != nil {
		return err
	}

	// Check if OTP is expired
	if time.Now().After(expiration) {
		return domain.ErrOTPExpired
	}

	// Check if OTP matches
	if storedOTP != otp {
		return domain.ErrInvalidOTP
	}

	// Verify user
	user, err := s.userRepo.GetUser(ctx, phoneNumber)
	if err != nil {
		return err
	}
	user.Verify()
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Delete OTP after verification
	if err := s.otpRepo.DeleteOTP(ctx, phoneNumber); err != nil {
		return err
	}

	// Log the verification activity
	activity := &domain.Activity{
		PhoneNumber: phoneNumber,
		Type:        domain.ActivityVerify,
		Timestamp:   time.Now(),
	}
	if err := s.activityRepo.RecordActivity(ctx, activity); err != nil {
		return err
	}

	return nil
}

// LoginInitiate initiates the login process by sending an OTP to the user's phone number.
func (s *AuthService) LoginInitiate(ctx context.Context, phoneNumber string) error {
	user, err := s.userRepo.GetUser(ctx, phoneNumber)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if !user.IsVerified() {
		return domain.ErrUserNotVerified
	}

	// Request OTP for login
	if err := s.requestNewOTP(ctx, phoneNumber); err != nil {
		return err
	}

	// Log the login initiation activity
	activity := &domain.Activity{
		PhoneNumber: phoneNumber,
		Type:        domain.ActivityLogin,
		Timestamp:   time.Now(),
	}
	if err := s.activityRepo.RecordActivity(ctx, activity); err != nil {
		return err
	}

	return nil
}

// LoginWithPhoneNumberAndOTP verifies the OTP and logs the user in.
func (s *AuthService) ValidatePhoneNumberLogin(ctx context.Context, phoneNumber, otp string) error {
	// Retrieve OTP from the database
	storedOTP, expiration, err := s.otpRepo.GetOTP(ctx, phoneNumber)
	if err != nil {
		return err
	}

	// Check if OTP is expired
	if time.Now().After(expiration) {
		return domain.ErrOTPExpired
	}

	// Check if OTP matches
	if storedOTP != otp {
		return domain.ErrInvalidOTP
	}

	// Verify user
	user, err := s.userRepo.GetUser(ctx, phoneNumber)
	if err != nil {
		return err
	}

	if !user.IsVerified() {
		return domain.ErrUserNotVerified
	}

	// Delete OTP after verification
	if err := s.otpRepo.DeleteOTP(ctx, phoneNumber); err != nil {
		return err
	}

	// Log the login activity
	activity := &domain.Activity{
		PhoneNumber: phoneNumber,
		Type:        domain.ActivityLogin,
		Timestamp:   time.Now(),
	}
	if err := s.activityRepo.RecordActivity(ctx, activity); err != nil {
		return err
	}

	return nil
}

// GetProfile retrieves the profile information of the user.
func (s *AuthService) GetProfile(ctx context.Context, phoneNumber string) (*domain.User, error) {
	user, err := s.userRepo.GetUser(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}

	// Log the profile retrieval activity
	activity := &domain.Activity{
		PhoneNumber: phoneNumber,
		Type:        domain.ActivityUpdate,
		Timestamp:   time.Now(),
	}
	if err := s.activityRepo.RecordActivity(ctx, activity); err != nil {
		return nil, err
	}

	return user, nil
}
