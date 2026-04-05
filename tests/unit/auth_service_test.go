package unit

import (
	"context"
	"testing"

	"github.com/Yashi213/room-booking/internal/auth"
	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/Yashi213/room-booking/internal/mocks"
	"github.com/Yashi213/room-booking/internal/repository"
	"github.com/Yashi213/room-booking/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type stubIssuer struct {
	token string
	err   error
}

func (s *stubIssuer) SignAccessToken(userID uuid.UUID, role string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}

var _ auth.AccessTokenIssuer = (*stubIssuer)(nil)

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	svc := service.NewAuthService(mocks.NewUsersRepository(t), &stubIssuer{token: "t"})
	_, err := svc.Register(context.Background(), domain.UserRegisterInput{
		Email:    "  ",
		Password: "12345678",
		Role:     domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrInvalidRegisterInput)
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	svc := service.NewAuthService(mocks.NewUsersRepository(t), &stubIssuer{})
	_, err := svc.Register(context.Background(), domain.UserRegisterInput{
		Email:    "a@b.c",
		Password: "short",
		Role:     domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrInvalidRegisterInput)
}

func TestAuthService_Register_InvalidRole(t *testing.T) {
	svc := service.NewAuthService(mocks.NewUsersRepository(t), &stubIssuer{})
	_, err := svc.Register(context.Background(), domain.UserRegisterInput{
		Email:    "a@b.c",
		Password: "12345678",
		Role:     domain.UserRole("god"),
	})
	require.ErrorIs(t, err, service.ErrInvalidRegisterInput)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := mocks.NewUsersRepository(t)
	repo.EXPECT().CreateUser(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(domain.User{}, repository.ErrConflict)

	svc := service.NewAuthService(repo, &stubIssuer{})
	_, err := svc.Register(context.Background(), domain.UserRegisterInput{
		Email:    "dup@x.y",
		Password: "12345678",
		Role:     domain.RoleUser,
	})
	require.ErrorIs(t, err, service.ErrEmailAlreadyExists)
}

func TestAuthService_Register_Success(t *testing.T) {
	uid := uuid.New()
	repo := mocks.NewUsersRepository(t)
	repo.EXPECT().CreateUser(mock.Anything, mock.Anything, "new@x.y", mock.Anything, domain.RoleUser).
		Return(domain.User{ID: uid, Email: "new@x.y", Role: domain.RoleUser}, nil)

	svc := service.NewAuthService(repo, &stubIssuer{})
	u, err := svc.Register(context.Background(), domain.UserRegisterInput{
		Email:    "New@X.Y",
		Password: "12345678",
		Role:     domain.RoleUser,
	})
	require.NoError(t, err)
	require.Equal(t, uid, u.ID)
	require.Equal(t, "new@x.y", u.Email)
}

func TestAuthService_Login_EmptyCredentials(t *testing.T) {
	svc := service.NewAuthService(mocks.NewUsersRepository(t), &stubIssuer{})
	_, err := svc.Login(context.Background(), domain.UserLoginInput{Email: "", Password: ""})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := mocks.NewUsersRepository(t)
	repo.EXPECT().GetUserWithPasswordByEmail(mock.Anything, "nope@x.y").Return(domain.User{}, "", repository.ErrNotFound)

	svc := service.NewAuthService(repo, &stubIssuer{})
	_, err := svc.Login(context.Background(), domain.UserLoginInput{Email: "nope@x.y", Password: "12345678"})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("rightone"), bcrypt.DefaultCost)
	require.NoError(t, err)
	uid := uuid.New()

	repo := mocks.NewUsersRepository(t)
	repo.EXPECT().GetUserWithPasswordByEmail(mock.Anything, "u@x.y").
		Return(domain.User{ID: uid, Role: domain.RoleUser}, string(hash), nil)

	svc := service.NewAuthService(repo, &stubIssuer{token: "jwt"})
	_, err = svc.Login(context.Background(), domain.UserLoginInput{Email: "u@x.y", Password: "wrongpass"})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_Success(t *testing.T) {
	pass := "correcthorse"
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	require.NoError(t, err)
	uid := uuid.New()

	repo := mocks.NewUsersRepository(t)
	repo.EXPECT().GetUserWithPasswordByEmail(mock.Anything, "ok@x.y").
		Return(domain.User{ID: uid, Role: domain.RoleUser}, string(hash), nil)

	svc := service.NewAuthService(repo, &stubIssuer{token: "signed-jwt"})
	token, err := svc.Login(context.Background(), domain.UserLoginInput{Email: "Ok@X.Y", Password: pass})
	require.NoError(t, err)
	require.Equal(t, "signed-jwt", token)
}
