package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Yashi213/room-booking/internal/auth"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, input domain.UserRegisterInput) (domain.User, error)
	Login(ctx context.Context, input domain.UserLoginInput) (token string, err error)
}

type authService struct {
	usersRepo   repository.UsersRepository
	tokenIssuer auth.AccessTokenIssuer
}

func NewAuthService(usersRepo repository.UsersRepository, issuer auth.AccessTokenIssuer) AuthService {
	return &authService{
		usersRepo:   usersRepo,
		tokenIssuer: issuer,
	}
}

func (s *authService) Register(ctx context.Context, input domain.UserRegisterInput) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password
	role := input.Role

	if email == "" || len(password) < 8 {
		return domain.User{}, ErrInvalidRegisterInput
	}
	if role != domain.RoleAdmin && role != domain.RoleUser {
		return domain.User{}, ErrInvalidRegisterInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	id := uuid.New()
	u, err := s.usersRepo.CreateUser(ctx, id, email, string(hash), role)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return domain.User{}, ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}
	return u, nil
}

func (s *authService) Login(ctx context.Context, input domain.UserLoginInput) (string, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password

	if email == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	u, hash, err := s.usersRepo.GetUserWithPasswordByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokenIssuer.SignAccessToken(u.ID, string(u.Role))
}
