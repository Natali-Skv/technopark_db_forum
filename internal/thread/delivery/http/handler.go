package handler

import (
	"net/http"
	"strconv"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	threadRepo "github.com/Natali-Skv/technopark_db_forum/internal/thread"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
)

const (
	SlugCtxKey     = "slug"
	SlugOrIdCtxKey = "slug"
)

type Handler struct {
	Repo threadRepo.Repo
}

func NewHandler(repo threadRepo.Repo) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) CreateThread(ctx echo.Context) error {
	thread := &models.Thread{}
	if err := ctx.Bind(thread); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}
	thread.ForumSlug = ctx.Param(SlugCtxKey)
	newThread, err := h.Repo.Create(thread)
	if err != nil {
		pgerr, _ := err.(pgx.PgError)
		switch pgerr.Code {
		case "23505":
			conflictForum, err := h.Repo.GetBySlugOrId(thread.Slug, 0)
			if err != nil || conflictForum == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
			}
			return ctx.JSON(http.StatusConflict, conflictForum)
		case "AAAA1":
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+thread.AuthorNick)
		case "AAAA3":
			return echo.NewHTTPError(http.StatusNotFound, errors.NO_THREAD_FORUM+thread.ForumSlug)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusCreated, newThread)
}

func (h *Handler) UpdateThread(ctx echo.Context) error {
	threadSlugOrId := ctx.Param(SlugOrIdCtxKey)
	threadId, err := strconv.Atoi(threadSlugOrId)
	thread := &models.Thread{}
	if err := ctx.Bind(thread); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}

	thread.Slug = threadSlugOrId
	thread.Id = threadId

	threadResp, err := h.Repo.UpdateThread(thread)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+threadSlugOrId)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, threadResp)
}

func (h *Handler) GetThread(ctx echo.Context) error {
	threadSlugOrId := ctx.Param(SlugOrIdCtxKey)
	threadId, err := strconv.Atoi(threadSlugOrId)
	threadResp, err := h.Repo.GetBySlugOrId(threadSlugOrId, int(threadId))
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_USER_BY_NICK+threadSlugOrId)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, threadResp)
}

func (h *Handler) Vote(ctx echo.Context) error {
	vote := &models.Vote{}
	if err := ctx.Bind(&vote); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}
	threadSlugOrId := ctx.Param(SlugOrIdCtxKey)
	ThreadId, err := strconv.Atoi(threadSlugOrId)
	if err == nil {
		vote.ThreadId = int(ThreadId)
	} else {
		vote.ThreadSlug = threadSlugOrId
	}
	thread, err := h.Repo.Vote(vote)
	if err != nil {
		if pgerr, converted := err.(pgx.PgError); converted {
			if pgerr.Code == "23502" || pgerr.Code == "23503" {
				return echo.NewHTTPError(http.StatusNotFound, errors.NO_THREAD+vote.ThreadSlug+strconv.Itoa(int(vote.ThreadId)))
			}
			if pgerr.Code == "AAAA1" {
				return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_POST_AUTHOR_BY_NICK+vote.Nick)
			}
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
		}
	}
	return ctx.JSON(http.StatusOK, thread)
}
