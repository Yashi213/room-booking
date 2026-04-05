package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type AuthUser struct {
	ID   uuid.UUID
	Role UserRole
}

type User struct {
	ID        uuid.UUID
	Email     string
	Role      UserRole
	CreatedAt time.Time
}

type UserRegisterInput struct {
	Email    string
	Password string
	Role     UserRole
}

type UserLoginInput struct {
	Email    string
	Password string
}
