package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"CquptFunAnnihilator/client"
	"CquptFunAnnihilator/logger"
	"CquptFunAnnihilator/models"
)

type Authenticator struct {
	httpClient *client.HttpClient
	email      string
	password   string
	token      string
}

func NewAuthenticator(httpClient *client.HttpClient, email, password string) *Authenticator {
	logger.Info("Authenticator created",
		zap.String("email", email),
	)
	return &Authenticator{
		httpClient: httpClient,
		email:      email,
		password:   password,
	}
}

func (a *Authenticator) Login() error {
	logger.Info("Attempting login to platform",
		zap.String("email", a.email),
	)

	loginReq := models.LoginRequest{
		Email:    a.email,
		Password: a.password,
	}

	logger.Debug("Sending login request",
		zap.String("email", a.email),
	)

	resp, err := a.httpClient.Post("/login", loginReq)
	if err != nil {
		logger.Error("Login request failed",
			zap.String("email", a.email),
			zap.Error(err),
		)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Error("Login failed with non-OK status",
			zap.String("email", a.email),
			zap.Int("statusCode", resp.StatusCode()),
		)
		return fmt.Errorf("login failed with status: %d", resp.StatusCode())
	}

	var loginResp models.LoginResponse
	if err := json.Unmarshal(resp.Body(), &loginResp); err != nil {
		logger.Error("Failed to unmarshal login response",
			zap.Error(err),
		)
		return err
	}

	if loginResp.Code != 0 {
		logger.Error("Login failed with error code",
			zap.String("email", a.email),
			zap.Int("code", loginResp.Code),
			zap.String("message", loginResp.Message),
		)
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	a.token = loginResp.Data.Token
	a.httpClient.SetToken(a.token)

	logger.Info("Login successful",
		zap.String("email", a.email),
		zap.Int64("expiresAt", loginResp.Data.ExpiresAt),
	)

	return nil
}

func (a *Authenticator) GetToken() string {
	return a.token
}

func (a *Authenticator) IsLoggedIn() bool {
	return a.token != ""
}
