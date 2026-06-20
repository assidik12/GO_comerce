package service

import (
	"context"
	"fmt"
	"time"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/pkg/hash"
	"github.com/assidik12/catalyst/internal/pkg/jwt"
	"github.com/go-playground/validator/v10"

	"golang.org/x/crypto/bcrypt"
)

// UserService defines the business-logic contract for users.
type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
}

type userService struct {
	repo      domain.UserRepository
	validate  *validator.Validate
	jwtSecret string // Injected secret
}

// NewUserService constructs a UserService with its dependencies.
func NewUserService(
	repo domain.UserRepository,
	validate *validator.Validate,
	jwtSecret string, // Injected parameter
) UserService {
	return &userService{
		repo:      repo,
		validate:  validate,
		jwtSecret: jwtSecret,
	}
}

// Register implements UserService.
func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return dto.UserResponse{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	hashedPassword, err := hash.NewCryptoHasher(bcrypt.DefaultCost).HashPassword(req.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	newUser := domain.User{
		Name:      req.Name,
		Email:     req.Email,
		Role:      "user",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := s.repo.Save(ctx, newUser)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// Login implements UserService using the injected jwtSecret.
func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: email %s", domain.ErrNotFound, req.Email)
	}

	if err := hash.NewCryptoHasher(bcrypt.DefaultCost).ComparePassword(user.Password, req.Password); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: invalid email or password", domain.ErrUnauthorized)
	}

	// Use the injected field instead of global config
	token, err := jwt.NewJWTService(s.jwtSecret).GenerateJWT(user)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}
