package repo

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

func NewRepo(conn *pgx.ConnPool) *Repo {
	return &Repo{Conn: conn}
}
func (r *Repo) Create(forum *models.Forum) (*models.Forum, error) {
	err := r.Conn.QueryRow(`INSERT into forums(title, slug, author_nick) 
									VALUES ($1,$2,$3) RETURNING author_nick`, forum.Title, forum.Slug, forum.UserNick).Scan(&forum.UserNick)
	if err != nil {
		return nil, err
	}
	return forum, nil
}
func (r *Repo) GetBySlug(slug string) (*models.Forum, error) {
	forum := &models.Forum{}
	err := r.Conn.QueryRow(`SELECT slug, title, posts, threads, author_nick FROM forums WHERE slug =$1`, slug).Scan(&forum.Slug, &forum.Title, &forum.Posts, &forum.Threads, &forum.UserNick)
	if err != nil {
		return nil, err
	}
	return forum, nil
}
func (r *Repo) CheckBySlug(slug string) (bool, error) {
	var exists bool
	err := r.Conn.QueryRow(`Select exists(SELECT 1 FROM forums WHERE slug =$1)`, slug).Scan(&exists)
	return exists, err
}
func (r *Repo) GetForumThreads(slug string, desc bool, limit int, since string) ([]models.Thread, error) {
	args := make([]interface{}, 0, 3)
	query := `SELECT id, slug, title, author_nick, forum_slug, message, votes, created FROM threads WHERE forum_slug =$1 `
	args = append(args, slug)
	nextPlaceholderNum := 2
	if since != "" && desc {
		query += `AND created<=$` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, since)
		nextPlaceholderNum++
	}
	if since != "" && !desc {
		query += `AND created>=$` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, since)
		nextPlaceholderNum++
	}
	if desc {
		query += ` ORDER BY created DESC `
	} else {
		query += ` ORDER BY created ASC `
	}
	if limit != 0 {
		query += `LIMIT $` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, limit)
	}
	threadRows, err := r.Conn.Query(query, args...)
	defer threadRows.Close()
	if err != nil {
		return nil, err
	}
	threadsResp := make([]models.Thread, 0)
	for threadRows.Next() {
		thread := models.Thread{}
		var created time.Time
		var slug sql.NullString
		err = threadRows.Scan(&thread.Id, &slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
		if err != nil {
			return nil, err
		}
		thread.Created = strfmt.DateTime(created.UTC()).String()
		thread.Slug = slug.String
		threadsResp = append(threadsResp, thread)
	}
	return threadsResp, nil
}

func (r *Repo) GetForumUsers(slug string, desc bool, limit int, since string) ([]models.User, error) {
	args := make([]interface{}, 0, 4)
	args = append(args, slug)
	nextPlaceholderNum := 2

	query := `SELECT name,nick,email,about FROM forum_users fu JOIN users u ON fu.user_nick=u.nick WHERE forum_slug=$1`

	if desc {
		if since != "" {
			query += ` AND u.nick < $` + strconv.Itoa(nextPlaceholderNum)
			args = append(args, since)
			nextPlaceholderNum++
		}
		query += ` ORDER BY lower(nick) DESC`
	} else {
		if since != "" {
			query += ` AND u.nick > $` + strconv.Itoa(nextPlaceholderNum)
			args = append(args, since)
			nextPlaceholderNum++
		}
		query += ` ORDER BY u.nick`
	}

	if limit != 0 {
		query += ` LIMIT $` + strconv.Itoa(nextPlaceholderNum)
		nextPlaceholderNum++
		args = append(args, limit)
	}

	// SELECT name,nick,email,about FROM forum_users fu JOIN u ON fu.user_nick=u.nick WHERE forum_slug=$1
	// userRows, err := r.Conn.Query(`SELECT name,nick,email,about FROM (SELECT author_nick FROM threads WHERE forum_slug=$1`+sinceQuery+` UNION DISTINCT SELECT author_nick FROM posts WHERE forum_slug=$1`+sinceQuery+`) t JOIN users u ON t.author_nick=u.nick `+orderQuery, args...)
	userRows, err := r.Conn.Query(query, args...)
	defer userRows.Close()

	if err != nil {
		return nil, err
	}

	userResp := make([]models.User, 0)

	for userRows.Next() {
		user := models.User{}
		err = userRows.Scan(&user.Name, &user.Nick, &user.Email, &user.About)
		if err != nil {
			return nil, err
		}
		err = userRows.Err()
		if err != nil {
			return nil, err
		}
		userResp = append(userResp, user)
	}
	return userResp, nil
}
