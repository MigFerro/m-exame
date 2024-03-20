package services

import (
	"github.com/MigFerro/exame/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
)

type UserService struct {
	DB *sqlx.DB
}

func (s *UserService) UserExistsInDB(authId string) (entities.UserEntity, bool) {
	row := s.DB.QueryRowx("SELECT * FROM users where auth_id = $1", authId)

	var dbUser entities.UserEntity
	err := row.StructScan(&dbUser)

	return dbUser, err == nil
}

func (s *UserService) CreateUserFromGoth(gothUser goth.User) (entities.UserEntity, error) {
	tx := s.DB.MustBegin()
	tx.MustExec("INSERT INTO users (auth_id, email, name) VALUES ($1, $2, $3)",
		gothUser.UserID,
		gothUser.Email,
		gothUser.Name,
	)

	var dbUser entities.UserEntity
	row := tx.QueryRowx("SELECT * FROM users where auth_id = $1", gothUser.UserID)
	err := row.StructScan(&dbUser)
	tx.Commit()

	return dbUser, err
}

func (s *UserService) GetUserRole(userId uuid.UUID) entities.UserRoleEnum {
	row := s.DB.QueryRowx("SELECT * FROM users where id = $1", userId)

	var dbUser entities.UserEntity
	row.StructScan(&dbUser)

	return dbUser.Role
}
