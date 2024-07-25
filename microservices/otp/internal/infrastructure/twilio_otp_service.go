package infrastructure

import (
	"context"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

type TwilioOTPService struct {
	client     *twilio.RestClient
	serviceSID string
}

func NewTwilioOTPService(AccountSID, AuthToken, ServiceSID string) *TwilioOTPService {
	return &TwilioOTPService{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: AccountSID,
			Password: AuthToken,
		}),
		serviceSID: ServiceSID,
	}
}

func (s *TwilioOTPService) SendOTP(ctx context.Context, phone, code string) error {

	params := &verify.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel("sms")
	// params.SetCustomCode(code) // Premium Feature
	_, err := s.client.VerifyV2.CreateVerification(s.serviceSID, params)
	return err
}
