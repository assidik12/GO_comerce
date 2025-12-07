package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/assidik12/go-restfull-api/config"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/pkg/hash"
	"github.com/assidik12/go-restfull-api/internal/pkg/jwt"
	"github.com/assidik12/go-restfull-api/internal/repository/mysql"
	"github.com/go-playground/validator/v10"

	"golang.org/x/crypto/bcrypt"
)

// UserService mendefinisikan kontrak untuk logika bisnis user.
type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
}

type userService struct {
	repo     mysql.UserRepository
	DB       *sql.DB
	validate *validator.Validate
}

// Perhatikan return type-nya sekarang adalah interface
func NewUserService(repo mysql.UserRepository, DB *sql.DB, validate *validator.Validate) UserService {
	return &userService{
		repo:     repo,
		DB:       DB,
		validate: validate,
	}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	// Validasi input
	err := s.validate.Struct(req)
	if err != nil {
		return dto.UserResponse{}, errors.New("invalid input data")
	}
	// 1. Hash password
	hashedPassword, err := hash.NewCryptoHasher(bcrypt.DefaultCost).HashPassword(req.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	// 2. Buat objek domain User
	newUser := domain.User{
		Name:      req.Name,
		Email:     req.Email,
		Role:      "user",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. Simpan ke repository
	user, err := s.repo.Save(ctx, newUser)
	if err != nil {
		return dto.UserResponse{}, err
	}

	// 4. Kembalikan DTO response (tanpa password)
	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	cfg := config.GetConfig()

	// Validasi input
	err := s.validate.Struct(req)
	if err != nil {
		return dto.LoginResponse{}, errors.New("invalid input data")
	}

	// 1. Cari user berdasarkan email
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.LoginResponse{}, errors.New("user not found")
		}
		return dto.LoginResponse{}, err
	}

	// 2. Bandingkan password yang di-hash dengan password dari request
	err = hash.NewCryptoHasher(bcrypt.DefaultCost).ComparePassword(user.Password, req.Password)
	if err != nil {
		// Jika error, kemungkinan besar password tidak cocok
		return dto.LoginResponse{}, errors.New("invalid email or password")
	}

	// 3. Buat JWT token
	token, err := jwt.NewJWTService(cfg.JWTSecret).GenerateJWT(user)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	// 4. Kembalikan response dengan token
	return dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}
