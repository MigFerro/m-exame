package templateutils

import (
	"context"

	"github.com/MigFerro/exame/data"
)

func GetAuthenticatedUserInfo(c context.Context) *data.LoggedUser {
	authUser, ok := c.Value("authUser").(*data.LoggedUser)

	if !ok {
		return &data.LoggedUser{}
	}

	return authUser
}

func GetExameString(exameYear string, exameFase string) string {
	return "Exame Nacional de " + exameYear + ", " + exameFase + "ª fase"
}

func PrepLevelColor(prepLevel int) string {
	if prepLevel < 25 {
		return "text-red-700"
	}

	if prepLevel < 65 {
		return "text-yellow-500"
	}

	return "text-green-700"
}
