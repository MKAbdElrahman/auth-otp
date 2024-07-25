package application

import (
	"context"
	"log"

	"midaslabs/microservices/otp/internal/domain"
)

type OTPService struct {
	messageBroker domain.MessageBroker
	otpClient     domain.OTPServiceClient
}

func NewOTPService(messageBroker domain.MessageBroker, otpClient domain.OTPServiceClient) *OTPService {
	return &OTPService{
		messageBroker: messageBroker,
		otpClient:     otpClient,
	}
}

func (s *OTPService) Start(ctx context.Context) error {
	return s.messageBroker.Subscribe(ctx, "verification", s.handleOTPEvent)
}

func (s *OTPService) handleOTPEvent(ctx context.Context, message []byte) {
	event, err := domain.DeserializeOTPEvent(message)
	if err != nil {
		log.Printf("Failed to deserialize OTP event: %v", err)
		return
	}

	// Here you would send the OTP to the user via SMS, email, etc.
	// This part is left out for simplicity, but typically you'd integrate with an external service like Twilio.
	log.Printf("Sending OTP to phone number %s: %s", event.PhoneNumber, event.OTPCode)

	err = s.otpClient.SendOTP(ctx, event.PhoneNumber, event.OTPCode)

	if err != nil {
		log.Printf("Failed to send OTP: %v", err)
		return
	}

}
