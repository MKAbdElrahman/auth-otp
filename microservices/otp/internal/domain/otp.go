package domain

import (
	"context"
	"encoding/json"
)

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

type MessageBroker interface {
	Publish(ctx context.Context, topic string, message []byte) error
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, message []byte)) error
}

type OTPServiceClient interface {
	SendOTP(ctx context.Context, phone, code string) error
}
