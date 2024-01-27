package templateutils

import (
	"context"

	"github.com/MigFerro/exame/entities"
)

func GetAuthenticatedUserName(c context.Context) string {
	authUser, ok := c.Value("authUser").(*entities.AuthUser)
	if !ok {
		return ""
	}
	return authUser.Name
}
