package repo

import (
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
func (r *Repo) Create(thread *models.Thread) (*models.Thread, error) {
	var err error
	if thread.Created == "" {
		err = r.Conn.QueryRow(`INSERT into threads(slug, title, author_nick, forum_slug, message) 
		VALUES (NULLIF($1, ''),$2,$3,$4,$5) RETURNING author_nick, id, forum_slug`, thread.Slug, thread.Title, thread.AuthorNick, thread.ForumSlug, thread.Message).Scan(&thread.AuthorNick, &thread.Id, &thread.ForumSlug)
	} else {
		err = r.Conn.QueryRow(`INSERT into threads(slug, title, author_nick, forum_slug, message, created) 
		VALUES (NULLIF($1, ''),$2,$3,$4,$5,$6) RETURNING author_nick, id, forum_slug`, thread.Slug, thread.Title, thread.AuthorNick, thread.ForumSlug, thread.Message, thread.Created).Scan(&thread.AuthorNick, &thread.Id, &thread.ForumSlug)
	}
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (r *Repo) GetBySlugOrId(slug string, id int) (*models.Thread, error) {
	thread := &models.Thread{}
	var created time.Time
	err := r.Conn.QueryRow(`SELECT id, slug, title, author_nick, forum_slug, message, votes, created FROM threads WHERE slug =$1 OR id=$2`, slug, id).Scan(&thread.Id, &thread.Slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	thread.Created = strfmt.DateTime(created.UTC()).String()
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (r *Repo) UpdateThread(thread *models.Thread) (*models.Thread, error) {
	var created time.Time
	err := r.Conn.QueryRow(`UPDATE threads SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) WHERE $3!=0 AND id=$3 OR $4!='' AND slug=$4 RETURNING id, slug, title, author_nick, forum_slug, message, votes, created`, thread.Title, thread.Message, thread.Id, thread.Slug).Scan(&thread.Id, &thread.Slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	thread.Created = strfmt.DateTime(created.UTC()).String()
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (r *Repo) Vote(vote *models.Vote) (*models.Thread, error) {
	thread := &models.Thread{}
	var created time.Time
	// fmt.Println(`INSERT INTO votes(user_nick, thread_id, vote) VALUES (` + vote.Nick + `,(SELECT id FROM threads WHERE slug=` + vote.ThreadSlug + ` OR id=` + strconv.Itoa(int(vote.ThreadId)) + `),` + strconv.Itoa(vote.Voice) + `) ON CONFLICT(user_nick, thread_id) DO UPDATE SET vote=` + strconv.Itoa(vote.Voice))
	_, err := r.Conn.Exec(`INSERT INTO votes(user_nick, thread_id, vote) VALUES ($1,   			   (SELECT id FROM threads WHERE slug=$2 OR id=$3),$4) ON CONFLICT(user_nick, thread_id) DO UPDATE SET vote=$4`, vote.Nick, vote.ThreadSlug, vote.ThreadId, vote.Voice)
	if err != nil {
		return nil, err
	}
	err = r.Conn.QueryRow(`SELECT id, slug, title, author_nick, forum_slug, message, votes, created FROM threads WHERE slug = $1 OR id = $2`, vote.ThreadSlug, vote.ThreadId).Scan(&thread.Id, &thread.Slug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &created)
	thread.Created = strfmt.DateTime(created.UTC()).String()
	if err != nil {
		return nil, err
	}
	return thread, nil
}
