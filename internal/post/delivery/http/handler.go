package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	postRepo "github.com/Natali-Skv/technopark_db_forum/internal/post"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
)

const (
	SlugOrIdCtxKey     = "slug"
	IdCtxKey           = "id"
	DescSortQueryParam = "desc"
	SinceQueryParam    = "since"
	LimitQueryParam    = "limit"
	SortQueryParam     = "sort"
	RelatedQueryParam  = "related"
)

type Handler struct {
	Repo postRepo.Repo
}

func NewHandler(repo postRepo.Repo) *Handler {
	return &Handler{Repo: repo}
}
func (h *Handler) CreatePost(ctx echo.Context) error {
	posts := []models.Post{}
	if err := ctx.Bind(&posts); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}
	// if len(posts) == 0 {
	// 	if !h.Repo.HasThreadBySlugOrId() {
	// 		return echo.NewHTTPError(http.StatusNotFound, errors.NO_THREAD+posts[0].ThreadSlug+strconv.Itoa(int(posts[0].ThreadId)))
	// 	}
	// 	return ctx.JSON(http.StatusCreated, posts)
	// }
	threadSlugOrId := ctx.Param(SlugOrIdCtxKey)
	threadId, _ := strconv.Atoi(threadSlugOrId)

	newPost, err := h.Repo.Create(threadSlugOrId, int(threadId), posts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NO_THREAD+threadSlugOrId+strconv.Itoa(int(threadId)))
		}
		if pgerr, converted := err.(pgx.PgError); converted {
			if pgerr.Code == "AAAA0" {
				return echo.NewHTTPError(http.StatusConflict, errors.NO_PARENT_POST)
			}
			if pgerr.Code == "AAAA1" {
				return echo.NewHTTPError(http.StatusNotFound, errors.NOT_FOUND_POST_AUTHOR_BY_NICK+posts[0].AuthorNick)
			}
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
		}
	}
	return ctx.JSON(http.StatusCreated, newPost)
}
func (h *Handler) GetThreadPosts(ctx echo.Context) error {
	threadSlugOrId := ctx.Param(SlugOrIdCtxKey)
	threadId, _ := strconv.Atoi(threadSlugOrId)
	since, _ := strconv.Atoi(ctx.QueryParam(SinceQueryParam))
	sort := ctx.QueryParam(SortQueryParam)
	descStr := ctx.QueryParam(DescSortQueryParam)
	desc := false
	if descStr == "true" {
		desc = true
	}
	limit, _ := strconv.Atoi(ctx.QueryParam(LimitQueryParam))
	posts, err := h.Repo.GetThreadPosts(threadSlugOrId, int(threadId), desc, limit, int(since), sort)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	if len(posts) == 0 {
		if exists, err := h.Repo.CheckThreadBySlugOrId(threadSlugOrId, int(threadId)); !exists && err == nil {
			return echo.NewHTTPError(http.StatusNotFound, errors.NO_THREAD+threadSlugOrId+strconv.Itoa(int(threadId)))
		}
	}
	return ctx.JSON(http.StatusOK, posts)
}

func (h *Handler) GetPost(ctx echo.Context) error {
	id, _ := strconv.Atoi(ctx.Param(IdCtxKey))
	related := ctx.QueryParam(RelatedQueryParam)
	relatedArray := strings.Split(related, ",")

	post, user, forum, thread, err := h.Repo.GetPostByIdRelated(int(id), relatedArray)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NO_POST+strconv.Itoa(id))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	if post == nil {
		return echo.NewHTTPError(http.StatusNotFound, errors.NO_POST+strconv.Itoa(id))
	}
	return ctx.JSON(http.StatusOK, struct {
		Post   *models.Post   `json:"post"`
		User   *models.User   `json:"author"`
		Forum  *models.Forum  `json:"forum"`
		Thread *models.Thread `json:"thread"`
	}{Post: post, User: user, Forum: forum, Thread: thread})
}

func (h *Handler) UpdatePost(ctx echo.Context) error {
	post := &models.Post{}
	post.Id, _ = strconv.Atoi(ctx.Param(IdCtxKey))
	if err := ctx.Bind(post); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.BAD_BODY)
	}

	postResp, err := h.Repo.UpdatePost(post)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, errors.NO_POST+strconv.Itoa(post.Id))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errors.INTERNAL_SERVER_ERROR)
	}
	return ctx.JSON(http.StatusOK, postResp)
}
