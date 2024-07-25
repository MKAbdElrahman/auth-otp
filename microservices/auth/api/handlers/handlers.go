package handlers

import (
	"context"
	"encoding/json"
	"midaslabs/microservices/auth/internal/application"
	"midaslabs/microservices/auth/internal/domain"
	"net/http"

	"github.com/charmbracelet/log"
)

type AuthHandler struct {
	authService *application.AuthService
	logger      *log.Logger
}

func NewAuthHandler(logger *log.Logger, authService *application.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) SignUpWithPhoneNumber(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Handler: SignUpWithPhoneNumber: failed to decode request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.authService.SignUpWithPhoneNumber(context.Background(), request.PhoneNumber); err != nil {
		h.logger.Errorf("Handler: SignUpWithPhoneNumber: failed to sign up for phone number %s: %v", request.PhoneNumber, err)
		if err == domain.ErrUserAlreadyExists {
			http.Error(w, "User already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to sign up", http.StatusInternalServerError)
		}
		return
	}

	log.Infof("Handler: SignUpWithPhoneNumber: OTP sent to phone number %s", request.PhoneNumber)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP sent"))
}

func (h *AuthHandler) VerifyPhoneNumber(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phone"`
		OTP         string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Handler: VerifyPhoneNumber: failed to decode request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.authService.VerifyPhoneNumber(context.Background(), request.PhoneNumber, request.OTP); err != nil {
		h.logger.Errorf("Handler: VerifyPhoneNumber: failed to verify phone number %s: %v", request.PhoneNumber, err)
		http.Error(w, "Failed to verify phone number", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("Handler: VerifyPhoneNumber: phone number %s verified", request.PhoneNumber)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Phone number verified"))
}

func (h *AuthHandler) LoginInitiate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Handler: LoginInitiate: failed to decode request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := h.authService.LoginInitiate(context.Background(), request.PhoneNumber)
	if err != nil {
		h.logger.Errorf("Handler: LoginInitiate: failed to initiate login for phone number %s: %v", request.PhoneNumber, err)
		if err == domain.ErrUserNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to initiate login", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Infof("Handler: LoginInitiate: OTP sent to phone number %s", request.PhoneNumber)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP sent"))
}

func (h *AuthHandler) ValidatePhoneNumberLogin(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phone"`
		OTP         string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Handler: LoginWithPhoneNumberAndOTP: failed to decode request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.authService.ValidatePhoneNumberLogin(context.Background(), request.PhoneNumber, request.OTP); err != nil {
		h.logger.Errorf("Handler: LoginWithPhoneNumberAndOTP: failed to login with phone number %s: %v", request.PhoneNumber, err)
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("Handler: LoginWithPhoneNumberAndOTP: user %s logged in successfully", request.PhoneNumber)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Errorf("Handler: GetProfile: failed to decode request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	profile, err := h.authService.GetProfile(context.Background(), request.PhoneNumber)
	if err != nil {
		h.logger.Errorf("Handler: GetProfile: failed to get profile for phone number %s: %v", request.PhoneNumber, err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("Handler: GetProfile: retrieved profile for phone number %s", request.PhoneNumber)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}
