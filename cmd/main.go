package main

import (
	"fmt"
	"log"

	"github.com/Natali-Skv/technopark_db_forum/config"
	"github.com/Natali-Skv/technopark_db_forum/configRouting"
	forumHandler "github.com/Natali-Skv/technopark_db_forum/internal/forum/delivery/http"
	forumRepository "github.com/Natali-Skv/technopark_db_forum/internal/forum/repo"
	postHandler "github.com/Natali-Skv/technopark_db_forum/internal/post/delivery/http"
	postRepository "github.com/Natali-Skv/technopark_db_forum/internal/post/repo"
	serviceHandler "github.com/Natali-Skv/technopark_db_forum/internal/service/delivery/http"
	serviceRepository "github.com/Natali-Skv/technopark_db_forum/internal/service/repo"
	threadHandler "github.com/Natali-Skv/technopark_db_forum/internal/thread/delivery/http"
	threadRepository "github.com/Natali-Skv/technopark_db_forum/internal/thread/repo"
	userHandler "github.com/Natali-Skv/technopark_db_forum/internal/user/delivery/http"
	userRepository "github.com/Natali-Skv/technopark_db_forum/internal/user/repo"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s",
		config.DbConfig.User, config.DbConfig.Password, config.DbConfig.DBName, config.DbConfig.Port)
	pgxConn, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		log.Fatal(err.Error())
	}
	pgxConn.PreferSimpleProtocol = true
	config := pgx.ConnPoolConfig{
		ConnConfig:     pgxConn,
		MaxConnections: config.DbConfig.MaxConnections,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}
	connPool, err := pgx.NewConnPool(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	e := echo.New()
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	userRepo := userRepository.NewRepo(connPool)
	userHandler := userHandler.NewHandler(userRepo)
	forumRepo := forumRepository.NewRepo(connPool)
	forumHandler := forumHandler.NewHandler(forumRepo)
	threadRepo := threadRepository.NewRepo(connPool)
	threadHandler := threadHandler.NewHandler(threadRepo)
	postRepo := postRepository.NewRepo(connPool)
	postHandler := postHandler.NewHandler(postRepo)
	servRepo := serviceRepository.NewRepo(connPool)
	servHandler := serviceHandler.NewHandler(servRepo)

	handlers := configRouting.Handlers{
		UserHandler:    userHandler,
		ForumHandler:   forumHandler,
		ThreadHandler:  threadHandler,
		PostHandler:    postHandler,
		ServiceHandler: servHandler,
	}
	handlers.ConfigureRouting(e)
	e.Logger.Fatal(e.Start(":5000"))
}
