package templateutils

import (
	"context"

	"github.com/MigFerro/exame/data"
)

func GetAuthenticatedUserName(c context.Context) string {
	authUser, ok := c.Value("authUser").(*data.LoggedUser)
	if !ok {
		return ""
	}
	return authUser.Name
}
