package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (

	// --- DATABASE ERRORS
	ErrUserNotFound = errors.New("user not found")
	ErrOTPNotFound  = errors.New("OTP not found")
	// -- APPLICATION ERRORS
	ErrOTPExpired          = errors.New("OTP expired")
	ErrInvalidOTP          = errors.New("invalid OTP")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotVerified     = errors.New("user not verified")
	ErrUserAlreadyVerified = errors.New("user already verified")
)

type UserRepository interface {
	GetUser(ctx context.Context, phoneNumber string) (*User, error)
	AddUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
}

type User struct {
	PhoneNumber string
	Verified    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUser(phoneNumber string) *User {
	return &User{
		PhoneNumber: phoneNumber,
		Verified:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (u *User) Verify() {
	u.Verified = true
	u.UpdatedAt = time.Now()
}

func (u *User) IsVerified() bool {
	return u.Verified
}

type OTP struct {
	PhoneNumber string
	Code        string
	Expiration  time.Time
}
type OTPRepository interface {
	StoreOTP(ctx context.Context, phoneNumber, otp string, expiration time.Time) error
	GetOTP(ctx context.Context, phoneNumber string) (otp string, expiration time.Time, err error)
	DeleteOTP(ctx context.Context, phoneNumber string) error
}

type Activity struct {
	PhoneNumber string
	Type        ActivityType
	Timestamp   time.Time
}

type ActivityType string

const (
	ActivityLogin  ActivityType = "login"
	ActivityLogout ActivityType = "logout"
	ActivitySignup ActivityType = "signup"
	ActivityVerify ActivityType = "verify"
	ActivityUpdate ActivityType = "update"
	ActivityDelete ActivityType = "delete"
)

type ActivityRepository interface {
	RecordActivity(ctx context.Context, activity *Activity) error
}

type MessageBroker interface {
	Publish(ctx context.Context, topic string, message []byte) error
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, message []byte)) error
}

type OTPVerificationEvent struct {
	PhoneNumber string `json:"phoneNumber"`
	OTPCode     string `json:"otpCode"`
}

func (e *OTPVerificationEvent) Serialize() ([]byte, error) {
	return json.Marshal(e)

}

func DeserializeOTPEvent(data []byte) (*OTPVerificationEvent, error) {
	var event OTPVerificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
