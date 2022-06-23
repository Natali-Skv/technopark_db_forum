package handler

import (
	"net/http"
	"sync"

	"github.com/Natali-Skv/technopark_db_forum/internal/service"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Repo service.Repo
}

var Statistic = map[string]uint64{
	"status": 0,
	"clean":  0,
}
var StatisticMutex = sync.RWMutex{}

func NewHandler(repo service.Repo) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) Status(ctx echo.Context) error {
	StatisticMutex.Lock()
	Statistic["status"]++
	StatisticMutex.Unlock()
	status, err := h.Repo.Status()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, status)
}

func (h *Handler) ClearDB(ctx echo.Context) error {
	StatisticMutex.Lock()
	Statistic["clean"]++
	StatisticMutex.Unlock()
	err := h.Repo.TruncateDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.NoContent(http.StatusOK)
}
