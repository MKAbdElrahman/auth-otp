package infrastructure

import (
	"context"
	"fmt"

	"github.com/twilio/twilio-go"
)

type MockOTPService struct {
	client     *twilio.RestClient
	serviceSID string
}

func NewMockOTPService(AccountSID, AuthToken, ServiceSID string) *MockOTPService {
	return &MockOTPService{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: AccountSID,
			Password: AuthToken,
		}),
		serviceSID: ServiceSID,
	}
}

func (s *MockOTPService) SendOTP(ctx context.Context, phone, code string) error {
	fmt.Printf("Mock: Sent OTP %s to phone %s\n", code, phone)
	return nil
}
