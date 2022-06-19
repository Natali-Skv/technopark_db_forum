package handler

import (
	"net/http"

	"github.com/Natali-Skv/technopark_db_forum/internal/service"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Repo service.Repo
}

func NewHandler(repo service.Repo) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) Status(ctx echo.Context) error {
	status, err := h.Repo.Status()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, status)
}

func (h *Handler) ClearDB(ctx echo.Context) error {
	err := h.Repo.TruncateDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.NoContent(http.StatusOK)
}
