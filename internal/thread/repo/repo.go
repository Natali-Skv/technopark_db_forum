package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

func NewRepo(conn *pgx.ConnPool) *Repo {
	conn.Prepare("create_thread_now", "INSERT into threads(slug, title, author_nick, forum_slug, message) VALUES (NULLIF($1, ''),$2,$3,$4,$5) RETURNING author_nick, id, forum_slug")
	conn.Prepare("create_thread", "INSERT into threads(slug, title, author_nick, forum_slug, message, created) VALUES (NULLIF($1, ''),$2,$3,$4,$5,$6) RETURNING author_nick, id, forum_slug")
	conn.Prepare("get_thread_by_slug", "SELECT id, slug, title, author_nick, forum_slug, message, votes, created FROM threads WHERE slug =$1")
	conn.Prepare("get_thread_by_id", "SELECT id, slug, title, author_nick, forum_slug, message, votes, created FROM threads WHERE id=$1")
	conn.Prepare("update_thread", "UPDATE threads SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) WHERE $3!=0 AND id=$4 OR $5!='' AND slug=$6 RETURNING id, slug, title, author_nick, forum_slug, message, votes, created")
	conn.Prepare("vote_thread_by_id", "INSERT INTO votes(user_nick, thread_id, vote) VALUES ($1,$2,$3) ON CONFLICT(user_nick, thread_id) DO UPDATE SET vote=$4")
	conn.Prepare("vote_thread_by_slug", "INSERT INTO votes(user_nick, thread_id, vote) VALUES ($1, (SELECT id FROM threads WHERE slug=$2),$3) ON CONFLICT(user_nick, thread_id) DO UPDATE SET vote=$4 RETURNING thread_id")
	return &Repo{Conn: conn}
}
func (r *Repo) Create(thread *models.Thread) (*models.Thread, error) {
	var err error
	if thread.Created == "" {
		err = r.Conn.QueryRow("EXECUTE create_thread_now($1,$2,$3,$4,$5)", thread.Slug, thread.Title, thread.AuthorNick, thread.ForumSlug, thread.Message).Scan(&thread.AuthorNick, &thread.Id, &thread.ForumSlug)
	} else {
		err = r.Conn.QueryRow("EXECUTE create_thread($1,$2,$3,$4,$5,$6)", thread.Slug, thread.Title, thread.AuthorNick, thread.ForumSlug, thread.Message, thread.Created).Scan(&thread.AuthorNick, &thread.Id, &thread.ForumSlug)
	}
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (r *Repo) GetBySlugOrId(slug string, id int) (*models.Thread, error) {
	thread := &models.Thread{}
	var created time.Time
	var threadSlug sql.NullString
	var err error
	if id != 0 {
		err = r.Conn.QueryRow("EXECUTE get_thread_by_id($1)", id).Scan(&thread.Id, &threadSlug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	} else {
		err = r.Conn.QueryRow("EXECUTE get_thread_by_slug($1)", slug).Scan(&thread.Id, &threadSlug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	}
	if err != nil {
		return nil, err
	}
	thread.Created = strfmt.DateTime(created.UTC()).String()
	thread.Slug = threadSlug.String
	return thread, nil
}

func (r *Repo) UpdateThread(thread *models.Thread) (*models.Thread, error) {
	var created time.Time
	var slug sql.NullString
	err := r.Conn.QueryRow("EXECUTE update_thread($1,$2,$3,$4,$5,$6)", thread.Title, thread.Message, thread.Id, thread.Id, thread.Slug, thread.Slug).Scan(&thread.Id, &slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	if err != nil {
		return nil, err
	}
	thread.Created = strfmt.DateTime(created.UTC()).String()
	thread.Slug = slug.String
	return thread, nil
}

func (r *Repo) Vote(vote *models.Vote) (*models.Thread, error) {
	thread := &models.Thread{}
	var created time.Time
	var err error
	if vote.ThreadId != 0 {
		_, err = r.Conn.Exec("EXECUTE vote_thread_by_id($1,$2,$3,$4)", vote.Nick, vote.ThreadId, vote.Voice, vote.Voice)
	} else {
		err = r.Conn.QueryRow("EXECUTE vote_thread_by_slug($1,$2,$3,$4)", vote.Nick, vote.ThreadSlug, vote.Voice, vote.Voice).Scan(&vote.ThreadId)
	}
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	var slug sql.NullString
	err = r.Conn.QueryRow("EXECUTE get_thread_by_id($1)", vote.ThreadId).Scan(&thread.Id, &slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	if err != nil {
		return nil, err
	}
	thread.Created = strfmt.DateTime(created.UTC()).String()
	thread.Slug = slug.String

	return thread, nil
}
