package entities

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserRoleEnum string

const (
	Student UserRoleEnum = "STUDENT"
	Prof    UserRoleEnum = "PROF"
	Admin   UserRoleEnum = "ADMIN"
)

type UserEntity struct {
	Id        uuid.UUID    `json:"id" db:"id"`
	AuthId    string       `json:"auth_id" db:"auth_id"`
	Name      string       `json:"name" db:"name"`
	Email     string       `json:"email" db:"email"`
	Role      UserRoleEnum `json:"role" db:"role"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
}

type AuthUser struct {
	AuthId string
	Id     uuid.UUID
	Name   string
	Email  string
}
