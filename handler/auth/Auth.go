package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Auth handles OIDC authentication flow
// GET /api/auth?Code=<code>
func (base Base) Auth(c *gin.Context, req *AuthRequest) *AuthResponse {
	// Get OIDC configuration
	tokenEndpoint := viper.GetString("oidc.token_endpoint")
	userinfoEndpoint := viper.GetString("oidc.userinfo_endpoint")
	clientID := viper.GetString("oidc.client_id")
	clientSecret := os.Getenv("SONORA_AUTH_PWD")
	redirectURI := viper.GetString("oidc.redirect_uri")

	if tokenEndpoint == "" || userinfoEndpoint == "" || clientID == "" || clientSecret == "" {
		handler.Errorf(c, "OIDC is not properly configured")
		return nil
	}

	// Step 1: Exchange authorization code for access token
	tokenResp, err := exchangeCodeForToken(tokenEndpoint, req.Code, clientID, clientSecret, redirectURI)
	if err != nil {
		handler.Errorf(c, "failed to exchange code for token: %v", err)
		return nil
	}

	// Step 2: Get user info using access token
	userInfo, err := getUserInfo(userinfoEndpoint, tokenResp.AccessToken)
	if err != nil {
		handler.Errorf(c, "failed to get user info: %v", err)
		return nil
	}

	// Step 3: Encrypt email and return as token
	encryptedToken, err := service.Encrypt(service.CryptKey, userInfo.Email)
	if err != nil {
		handler.Errorf(c, "failed to encrypt token: %v", err)
		return nil
	}

	return &AuthResponse{
		Token: encryptedToken,
	}
}

type AuthRequest struct {
	Code string `form:"Code" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

type UserInfoResponse struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

// exchangeCodeForToken exchanges the authorization code for an access token
func exchangeCodeForToken(tokenEndpoint, code, clientID, clientSecret, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// getUserInfo retrieves user information from the userinfo endpoint
func getUserInfo(userinfoEndpoint, accessToken string) (*UserInfoResponse, error) {
	req, err := http.NewRequest("GET", userinfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send userinfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}
