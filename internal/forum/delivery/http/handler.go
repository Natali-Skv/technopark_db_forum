package handler

import (
	"net/http"
	"strconv"
	"sync"

	forumRepo "github.com/Natali-Skv/technopark_db_forum/internal/forum"
	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
)

const (
	SlugCtxKey         = "slug"
	DescSortQueryParam = "desc"
	SinceQueryParam    = "since"
	LimitQueryParam    = "limit"
)

var Statistic = map[string]uint64{
	"create forum":      0,
	"get forum":         0,
	"get forum users":   0,
	"get forum threads": 0,
}
var StatisticMutex = sync.RWMutex{}

type Handler struct {
	Repo forumRepo.Repo
}

func NewHandler(repo forumRepo.Repo) *Handler {
	return &Handler{Repo: repo}
}
func (h *Handler) CreateForum(ctx echo.Context) error {
	defer func() {
		StatisticMutex.Lock()
		Statistic["create forum"]++
		StatisticMutex.Unlock()
	}()
	forum := &models.Forum{}
	if err := ctx.Bind(forum); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}
	newForum, err := h.Repo.Create(forum)
	if err != nil {
		pgerr, _ := err.(pgx.PgError)
		switch pgerr.Code {
		case "23505":
			conflictForum, err := h.Repo.GetBySlug(forum.Slug)
			if err != nil || conflictForum == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
			}
			return ctx.JSON(http.StatusConflict, conflictForum)
		case "AAAA1":
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+forum.UserNick)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusCreated, newForum)
}
func (h *Handler) GetForum(ctx echo.Context) error {
	defer func() {
		StatisticMutex.Lock()
		Statistic["get forum"]++
		StatisticMutex.Unlock()
	}()
	slug := ctx.Param(SlugCtxKey)
	userResp, err := h.Repo.GetBySlug(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+slug)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, userResp)
}

func (h *Handler) GetForumThreads(ctx echo.Context) error {
	defer func() {
		StatisticMutex.Lock()
		Statistic["get forum threads"]++
		StatisticMutex.Unlock()
	}()
	since := ctx.QueryParam(SinceQueryParam)
	descStr := ctx.QueryParam(DescSortQueryParam)
	desc := false
	if descStr == "true" {
		desc = true
	}
	limit, _ := strconv.Atoi(ctx.QueryParam(LimitQueryParam))
	slug := ctx.Param(SlugCtxKey)
	threads, err := h.Repo.GetForumThreads(slug, desc, limit, since)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	if len(threads) == 0 {
		if exists, err := h.Repo.CheckBySlug(slug); !exists && err == nil {
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+slug)
		}
	}
	return ctx.JSON(http.StatusOK, threads)
}

func (h *Handler) GetForumUsers(ctx echo.Context) error {
	defer func() {
		StatisticMutex.Lock()
		Statistic["get forum users"]++
		StatisticMutex.Unlock()
	}()
	slug := ctx.Param(SlugCtxKey)

	limit, _ := strconv.Atoi(ctx.QueryParam(LimitQueryParam))
	since := ctx.QueryParam(SinceQueryParam)

	descStr := ctx.QueryParam(DescSortQueryParam)
	desc := false
	if descStr == "true" {
		desc = true
	}

	users, err := h.Repo.GetForumUsers(slug, desc, limit, since)

	if err != nil {
		return ctx.JSON(http.StatusOK, users)
	}
	if len(users) == 0 {
		if exists, err := h.Repo.CheckBySlug(slug); !exists && err == nil {
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+slug)
		}
		return ctx.JSON(http.StatusOK, []models.User{})
	}

	return ctx.JSON(http.StatusOK, users)
}
