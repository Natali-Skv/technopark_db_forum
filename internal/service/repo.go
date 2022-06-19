package service

import "github.com/Natali-Skv/technopark_db_forum/internal/models"

type Repo interface {
	Status() (*models.Status, error)
	TruncateDB() error
}
