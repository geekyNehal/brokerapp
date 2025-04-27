package user

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) SignUp(ctx context.Context, req *SignUpRequest) (*TokenResponse, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && err != ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, err
	}
	log.Printf("Generated password hash: %s", string(hashedPassword))

	user := &User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	createdUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("Error fetching created user: %v", err)
		return nil, err
	}

	return s.generateTokens(ctx, createdUser.ID)
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	log.Printf("Attempting login for email: %s", req.Email)

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == ErrUserNotFound {
			log.Printf("Login failed: user not found for email %s", req.Email)
			return nil, ErrInvalidCredentials
		}
		log.Printf("Login failed: error fetching user: %v", err)
		return nil, err
	}

	log.Printf("Found user with ID: %d", user.ID)
	log.Printf("Stored password hash: %s", user.Password)
	log.Printf("Attempting password comparison")

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("Password comparison failed: %v", err)
		log.Printf("Input password length: %d", len(req.Password))
		log.Printf("Stored hash length: %d", len(user.Password))
		return nil, ErrInvalidCredentials
	}

	log.Printf("Login successful for user %s", user.Email)
	return s.generateTokens(ctx, user.ID)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	token, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if time.Now().After(token.ExpiresAt) {
		if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
			return nil, err
		}
		return nil, ErrTokenExpired
	}

	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, token.UserID)
}

func (s *Service) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) generateTokens(ctx context.Context, userID int64) (*TokenResponse, error) {
	accessTokenDuration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		accessTokenDuration = 5 * time.Minute
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(accessTokenDuration).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refreshTokenDuration, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		refreshTokenDuration = 24 * time.Hour
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(refreshTokenDuration).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	if err := s.repo.StoreRefreshToken(ctx, userID, refreshTokenString, time.Now().Add(refreshTokenDuration)); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
