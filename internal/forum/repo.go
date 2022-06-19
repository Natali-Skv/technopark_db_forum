package forum

import "github.com/Natali-Skv/technopark_db_forum/internal/models"

type Repo interface {
	Create(forum *models.Forum) (*models.Forum, error)
	GetBySlug(slug string) (*models.Forum, error)
	CheckBySlug(slug string) (bool, error)
	GetForumThreads(slug string, desc bool, limit int, since string) ([]models.Thread, error)
	GetForumUsers(slug string, desc bool, limit int, since string) ([]models.User, error)
}
