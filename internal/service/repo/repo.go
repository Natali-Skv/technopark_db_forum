package userRepo

import (
	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

func NewRepo(conn *pgx.ConnPool) *Repo {
	return &Repo{Conn: conn}
}

func (r *Repo) Status() (*models.Status, error) {
	status := &models.Status{}
	countRows, err := r.Conn.Query(`SELECT COUNT(*) FROM forums UNION ALL SELECT COUNT(*) FROM posts UNION ALL SELECT COUNT(*) FROM threads UNION ALL SELECT COUNT(*) FROM users`)
	if err != nil {
		return nil, err
	}

	countRows.Next()
	countRows.Scan(&status.Forums)
	countRows.Next()
	countRows.Scan(&status.Posts)
	countRows.Next()
	countRows.Scan(&status.Threads)
	countRows.Next()
	countRows.Scan(&status.Users)
	return status, nil
}

func (r *Repo) TruncateDB() error {
	_, err := r.Conn.Exec(`TRUNCATE users, forums, threads, posts, votes`)
	return err
}
