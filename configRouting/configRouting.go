package configRouting

import (
	forumHandler "github.com/Natali-Skv/technopark_db_forum/internal/forum/delivery/http"
	postHandler "github.com/Natali-Skv/technopark_db_forum/internal/post/delivery/http"
	serviceHandler "github.com/Natali-Skv/technopark_db_forum/internal/service/delivery/http"
	threadHandler "github.com/Natali-Skv/technopark_db_forum/internal/thread/delivery/http"
	userHandler "github.com/Natali-Skv/technopark_db_forum/internal/user/delivery/http"
	"github.com/labstack/echo/v4"
)

const (
	routerPrefix = "/api/"
)

type Handlers struct {
	UserHandler    *userHandler.Handler
	ForumHandler   *forumHandler.Handler
	ThreadHandler  *threadHandler.Handler
	PostHandler    *postHandler.Handler
	ServiceHandler *serviceHandler.Handler
}

func (hs *Handlers) ConfigureRouting(router *echo.Echo) {
	router.POST(routerPrefix+"user/:"+userHandler.NickCtxKey+"/create", hs.UserHandler.CreateUser)
	router.GET(routerPrefix+"user/:"+userHandler.NickCtxKey+"/profile", hs.UserHandler.GetUser)
	router.POST(routerPrefix+"user/:"+userHandler.NickCtxKey+"/profile", hs.UserHandler.UpdateUser)
	router.POST(routerPrefix+"forum/create", hs.ForumHandler.CreateForum)
	router.GET(routerPrefix+"forum/:"+forumHandler.SlugCtxKey+"/details", hs.ForumHandler.GetForum)
	router.GET(routerPrefix+"forum/:"+forumHandler.SlugCtxKey+"/threads", hs.ForumHandler.GetForumThreads)
	router.GET(routerPrefix+"forum/:"+forumHandler.SlugCtxKey+"/users", hs.ForumHandler.GetForumUsers)

	router.POST(routerPrefix+"forum/:"+threadHandler.SlugCtxKey+"/create", hs.ThreadHandler.CreateThread)
	router.POST(routerPrefix+"thread/:"+threadHandler.SlugCtxKey+"/vote", hs.ThreadHandler.Vote)
	router.GET(routerPrefix+"thread/:"+threadHandler.SlugCtxKey+"/details", hs.ThreadHandler.GetThread)
	router.POST(routerPrefix+"thread/:"+threadHandler.SlugCtxKey+"/details", hs.ThreadHandler.UpdateThread)

	router.POST(routerPrefix+"thread/:"+threadHandler.SlugCtxKey+"/create", hs.PostHandler.CreatePost)
	router.GET(routerPrefix+"thread/:"+threadHandler.SlugCtxKey+"/posts", hs.PostHandler.GetThreadPosts)
	router.GET(routerPrefix+"post/:"+postHandler.IdCtxKey+"/details", hs.PostHandler.GetPost)
	router.POST(routerPrefix+"post/:"+postHandler.IdCtxKey+"/details", hs.PostHandler.UpdatePost)

	router.GET(routerPrefix+"service/status", hs.ServiceHandler.Status)
	router.POST(routerPrefix+"service/clear", hs.ServiceHandler.ClearDB)
}
