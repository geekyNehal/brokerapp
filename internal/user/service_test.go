package user

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func TestSignUp(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	req := &SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// First call to GetUserByEmail should return not found
	mockRepo.On("GetUserByEmail", ctx, req.Email).Return(nil, ErrUserNotFound).Once()

	// CreateUser should be called with a new user
	mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(u *User) bool {
		return u.Email == req.Email && u.Password != ""
	})).Return(nil).Once()

	// Second call to GetUserByEmail should return the created user
	createdUser := &User{
		ID:    1,
		Email: req.Email,
	}
	mockRepo.On("GetUserByEmail", ctx, req.Email).Return(createdUser, nil).Once()

	// StoreRefreshToken should be called
	mockRepo.On("StoreRefreshToken", ctx, int64(1), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil).Once()

	resp, err := service.SignUp(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestSignUpUserExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	req := &SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	existingUser := &User{
		ID:    1,
		Email: req.Email,
	}

	mockRepo.On("GetUserByEmail", ctx, req.Email).Return(existingUser, nil)

	resp, err := service.SignUp(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrEmailAlreadyExists, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := hashPassword(password)

	user := &User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	req := &LoginRequest{
		Email:    user.Email,
		Password: password,
	}

	mockRepo.On("GetUserByEmail", ctx, req.Email).Return(user, nil)
	mockRepo.On("StoreRefreshToken", ctx, user.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	resp, err := service.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestLoginInvalidCredentials(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := hashPassword("different-password")

	user := &User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	req := &LoginRequest{
		Email:    user.Email,
		Password: password,
	}

	mockRepo.On("GetUserByEmail", ctx, req.Email).Return(user, nil)

	resp, err := service.Login(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	refreshToken := "valid-refresh-token"
	userID := int64(1)

	storedToken := &RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	mockRepo.On("GetRefreshToken", ctx, refreshToken).Return(storedToken, nil)
	mockRepo.On("DeleteRefreshToken", ctx, refreshToken).Return(nil)
	mockRepo.On("StoreRefreshToken", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	resp, err := service.RefreshToken(ctx, refreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestRefreshTokenExpired(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	ctx := context.Background()
	refreshToken := "expired-refresh-token"
	userID := int64(1)

	storedToken := &RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	mockRepo.On("GetRefreshToken", ctx, refreshToken).Return(storedToken, nil)
	mockRepo.On("DeleteRefreshToken", ctx, refreshToken).Return(nil)

	resp, err := service.RefreshToken(ctx, refreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrTokenExpired, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestService_Login(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	// Generate a valid bcrypt hash for "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		req           *LoginRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "Successful login",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				// Return existing user with correct password hash
				mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(&User{
					ID:       1,
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}, nil)
				// Store refresh token
				mockRepo.On("StoreRefreshToken", mock.Anything, int64(1), mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid credentials",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(&User{
					ID:       1,
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "User not found",
			req: &LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, ErrUserNotFound)
			},
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.Login(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_RefreshToken(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func()
		expectedError error
	}{
		{
			name:         "Successful token refresh",
			refreshToken: "valid-refresh-token",
			mockSetup: func() {
				// Return valid refresh token
				mockRepo.On("GetRefreshToken", mock.Anything, "valid-refresh-token").Return(&RefreshToken{
					UserID:    1,
					Token:     "valid-refresh-token",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
				// Delete old token
				mockRepo.On("DeleteRefreshToken", mock.Anything, "valid-refresh-token").Return(nil)
				// Store new token
				mockRepo.On("StoreRefreshToken", mock.Anything, int64(1), mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:         "Expired refresh token",
			refreshToken: "expired-token",
			mockSetup: func() {
				// Return expired refresh token
				mockRepo.On("GetRefreshToken", mock.Anything, "expired-token").Return(&RefreshToken{
					UserID:    1,
					Token:     "expired-token",
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}, nil)
				// Delete expired token
				mockRepo.On("DeleteRefreshToken", mock.Anything, "expired-token").Return(nil)
			},
			expectedError: ErrTokenExpired,
		},
		{
			name:         "Invalid refresh token",
			refreshToken: "invalid-token",
			mockSetup: func() {
				mockRepo.On("GetRefreshToken", mock.Anything, "invalid-token").Return(nil, ErrRefreshTokenNotFound)
			},
			expectedError: ErrRefreshTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := service.RefreshToken(context.Background(), tt.refreshToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	tests := []struct {
		name          string
		userID        int64
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "Successful user fetch",
			userID: 1,
			mockSetup: func() {
				mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(&User{
					ID:       1,
					Email:    "test@example.com",
					Password: "hashed_password",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "User not found",
			userID: 999,
			mockSetup: func() {
				mockRepo.On("GetUserByID", mock.Anything, int64(999)).Return(nil, ErrUserNotFound)
			},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_generateTokens(t *testing.T) {
	// Save original environment variables
	originalAccessDuration := os.Getenv("ACCESS_TOKEN_DURATION")
	originalRefreshDuration := os.Getenv("REFRESH_TOKEN_DURATION")
	defer func() {
		os.Setenv("ACCESS_TOKEN_DURATION", originalAccessDuration)
		os.Setenv("REFRESH_TOKEN_DURATION", originalRefreshDuration)
	}()

	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")

	tests := []struct {
		name               string
		accessDuration     string
		refreshDuration    string
		expectedAccessDur  time.Duration
		expectedRefreshDur time.Duration
	}{
		{
			name:               "Default durations",
			accessDuration:     "",
			refreshDuration:    "",
			expectedAccessDur:  5 * time.Minute,
			expectedRefreshDur: 24 * time.Hour,
		},
		{
			name:               "Custom durations",
			accessDuration:     "15m",
			refreshDuration:    "48h",
			expectedAccessDur:  15 * time.Minute,
			expectedRefreshDur: 48 * time.Hour,
		},
		{
			name:               "Invalid durations fallback to defaults",
			accessDuration:     "invalid",
			refreshDuration:    "invalid",
			expectedAccessDur:  5 * time.Minute,
			expectedRefreshDur: 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("ACCESS_TOKEN_DURATION", tt.accessDuration)
			os.Setenv("REFRESH_TOKEN_DURATION", tt.refreshDuration)

			// Setup mock
			mockRepo.On("StoreRefreshToken", mock.Anything, int64(1), mock.Anything, mock.Anything).Return(nil)

			// Generate tokens
			resp, err := service.generateTokens(context.Background(), 1)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)

			// Verify token durations
			accessToken, _ := jwt.Parse(resp.AccessToken, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})
			claims := accessToken.Claims.(jwt.MapClaims)
			exp := time.Unix(int64(claims["exp"].(float64)), 0)
			assert.WithinDuration(t, time.Now().Add(tt.expectedAccessDur), exp, time.Second)

			refreshToken, _ := jwt.Parse(resp.RefreshToken, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})
			claims = refreshToken.Claims.(jwt.MapClaims)
			exp = time.Unix(int64(claims["exp"].(float64)), 0)
			assert.WithinDuration(t, time.Now().Add(tt.expectedRefreshDur), exp, time.Second)

			mockRepo.AssertExpectations(t)
		})
	}
}
