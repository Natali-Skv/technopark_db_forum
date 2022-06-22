package userRepo

import (
	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

func NewRepo(conn *pgx.ConnPool) *Repo {
	conn.Prepare("status", " SELECT COUNT(*), sum(posts), sum(threads) FROM forums")
	conn.Prepare("status_users", " SELECT COUNT(*) FROM users")
	return &Repo{Conn: conn}
}

// просуммировать одновременно posts + threads
func (r *Repo) Status() (*models.Status, error) {
	status := &models.Status{}
	err := r.Conn.QueryRow("EXECUTE status").Scan(&status.Forums, &status.Posts, &status.Threads)
	err = r.Conn.QueryRow("EXECUTE status_users").Scan(&status.Users)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (r *Repo) TruncateDB() error {
	_, err := r.Conn.Exec(`TRUNCATE forum_users, users, forums, threads, posts, votes`)
	return err
}
