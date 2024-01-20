package handlers

import (
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
}

func (h HomeHandler) HomeShow(c echo.Context) error {
	return render(c, homeview.Show())
}
