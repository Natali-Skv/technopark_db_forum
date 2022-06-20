package repo

import (
	"database/sql"
	goErrors "errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

func NewRepo(conn *pgx.ConnPool) *Repo {
	return &Repo{Conn: conn}
}
func (r *Repo) Create(threadSlug string, threadId int, posts []models.Post) ([]models.Post, error) {
	var forumSlug string
	err := r.Conn.QueryRow(`SELECT forum_slug, id FROM threads WHERE slug=$1 OR id=$2`, threadSlug, threadId).Scan(&forumSlug, &threadId)
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return []models.Post{}, nil
	}
	query := `INSERT into posts(author_nick, parent_id,  message, forum_slug, thread_id) VALUES `
	fieldCount := 5
	args := make([]interface{}, 0, len(posts)*5)
	i := 0
	var post models.Post
	for i, post = range posts[:len(posts)-1] {
		query += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d),", i*fieldCount+1, i*fieldCount+2, i*fieldCount+3, i*fieldCount+4, i*fieldCount+5)
		args = append(args, post.AuthorNick, post.ParentId, post.Message, forumSlug, threadId)
		i += 1
	}
	post = posts[len(posts)-1]
	query += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d) RETURNING id, author_nick, created;", i*fieldCount+1, i*fieldCount+2, i*fieldCount+3, i*fieldCount+4, i*fieldCount+5)
	args = append(args, post.AuthorNick, post.ParentId, post.Message, forumSlug, threadId)
	postRows, err := r.Conn.Query(query, args...)
	defer postRows.Close()
	if err != nil {
		return nil, goErrors.New(errors.INTERNAL_SERVER_ERROR)
	}
	for i := range posts {
		postRows.Next()
		var created time.Time
		err := postRows.Err()
		if err != nil {
			return nil, err
		}
		scanErr := postRows.Scan(&posts[i].Id, &posts[i].AuthorNick, &created)
		posts[i].ForumSlug = forumSlug
		posts[i].ThreadId = threadId
		posts[i].Created = strfmt.DateTime(created.UTC()).String()
		if scanErr != nil {
			return nil, scanErr
		}
	}
	return posts, nil
}

func (r *Repo) GetThreadPosts(threadSlug string, threadId int, desc bool, limit int, since int, sort string) ([]models.Post, error) {
	switch sort {
	case "flat", "":
		return r.getThreadPostsFlat(threadSlug, threadId, desc, limit, since)
	case "tree":
		return r.getThreadPostsTree(threadSlug, threadId, desc, limit, since)
	case "parent_tree":
		return r.getThreadPostsParentTree(threadSlug, threadId, desc, limit, since)
	default:
		return nil, goErrors.New(errors.UNKNOWN_SORT_TYPE)
	}
}
func (r *Repo) getThreadPostsFlat(threadSlug string, threadId int, desc bool, limit int, since int) ([]models.Post, error) {
	args := make([]interface{}, 0, 3)
	query := `SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id = $1 OR ($2 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$2)) `
	args = append(args, threadId, threadSlug)
	nextPlaceholderNum := 3
	if since != 0 && desc {
		query += `AND id<$` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, since)
		nextPlaceholderNum++
	}
	if since != 0 && !desc {
		query += `AND id>$` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, since)
		nextPlaceholderNum++
	}
	if desc {
		query += ` ORDER BY created DESC,id DESC`
	} else {
		query += ` ORDER BY created,id ASC`
	}
	if limit != 0 {
		query += ` LIMIT $` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, limit)
	}
	threadRows, err := r.Conn.Query(query, args...)
	defer threadRows.Close()
	if err != nil {
		return nil, err
	}
	postsResp := make([]models.Post, 0)
	for threadRows.Next() {
		post := models.Post{}
		parentId := sql.NullInt64{}
		var created time.Time
		err = threadRows.Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
		post.ParentId = int(parentId.Int64)
		if err != nil {
			return nil, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		postsResp = append(postsResp, post)
	}
	return postsResp, nil
}
func (r *Repo) getThreadPostsTree(threadSlug string, threadId int, desc bool, limit int, since int) ([]models.Post, error) {
	args := make([]interface{}, 0, 3)
	query := `SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1 != 0 AND thread_id = $1 OR $2 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$2)) `
	args = append(args, threadId, threadSlug)
	nextPlaceholderNum := 3
	if since != 0 && desc {
		query += `AND path < (SELECT path FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)`
		args = append(args, since)
		nextPlaceholderNum++
	}
	if since != 0 && !desc {
		query += `AND path > (SELECT path FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)`
		args = append(args, since)
		nextPlaceholderNum++
	}
	if desc {
		query += ` ORDER BY path DESC `
	} else {
		query += ` ORDER BY path ASC `
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
	postsResp := make([]models.Post, 0)
	for threadRows.Next() {
		post := models.Post{}
		parentId := sql.NullInt64{}
		var created time.Time
		err = threadRows.Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
		post.ParentId = int(parentId.Int64)
		if err != nil {
			return nil, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		postsResp = append(postsResp, post)
	}
	return postsResp, nil
}
func (r *Repo) getThreadPostsParentTree(threadSlug string, threadId int, desc bool, limit int, since int) ([]models.Post, error) {
	args := make([]interface{}, 0, 3)
	var query string
	if limit == 0 {
		query = `SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1 != 0 AND thread_id = $1 OR $2 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$2)) `
		args = append(args, threadId, threadSlug)
		nextPlaceholderNum := 3
		if since != 0 && desc {
			query += `AND path < (SELECT path FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)`
			args = append(args, since)
			nextPlaceholderNum++
		}
		if since != 0 && !desc {
			query += `AND path > (SELECT path FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)`
			args = append(args, since)
			nextPlaceholderNum++
		}
		if desc {
			query += ` ORDER BY path[1] DESC, path, id`
		} else {
			query += ` ORDER BY path ASC`
		}
	} else {
		args = append(args, threadId, threadSlug)
		nextPlaceholderNum := 3

		query += `) t WHERE (dense_rank<=$` + strconv.Itoa(nextPlaceholderNum)
		args = append(args, limit)
		nextPlaceholderNum++
		if desc {
			if since != 0 {
				query = `AND path[1] < (SELECT path[1] FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)` + query
				args = append(args, since)
				nextPlaceholderNum++
			}
			query = `SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM 
					(SELECT id, parent_id, path, author_nick, forum_slug, thread_id, message, created, is_edited, dense_rank() OVER(ORDER BY path[1] DESC) FROM posts 
								WHERE ($1 != 0 AND thread_id = $1 OR $2 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$2)) ` + query + `) ORDER BY path[1] desc, path, id`

		} else {
			if since != 0 {
				query = `AND path[1] > (SELECT path[1] FROM posts WHERE id=$` + strconv.Itoa(nextPlaceholderNum) + `)` + query
				args = append(args, since)
				nextPlaceholderNum++
			}
			query = `SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM 
					(SELECT id, parent_id, path, author_nick, forum_slug, thread_id, message, created, is_edited, dense_rank() OVER(ORDER BY path[1]) FROM posts 
								WHERE ($1 != 0 AND thread_id = $1 OR $2 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$2)) ` + query + `) ORDER BY path ASC `
		}

	}

	threadRows, err := r.Conn.Query(query, args...)
	defer threadRows.Close()
	if err != nil {
		return nil, err
	}
	postsResp := make([]models.Post, 0)
	for threadRows.Next() {
		post := models.Post{}
		parentId := sql.NullInt64{}
		var created time.Time
		err = threadRows.Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
		post.ParentId = int(parentId.Int64)
		if err != nil {
			return nil, err
		}
		post.Created = strfmt.DateTime(created.UTC()).String()
		postsResp = append(postsResp, post)
	}
	return postsResp, nil
}

func (r *Repo) CheckThreadBySlugOrId(slug string, id int) (bool, error) {
	var exists bool
	err := r.Conn.QueryRow(`Select exists(SELECT 1 FROM threads WHERE slug =$1 OR id=$2)`, slug, id).Scan(&exists)
	return exists, err
}

func (r *Repo) GetPostByIdRelated(id int, related []string) (*models.Post, *models.User, *models.Forum, *models.Thread, error) {
	post := &models.Post{}
	var user *models.User
	var forum *models.Forum
	var thread *models.Thread

	var created time.Time
	var threadCreated time.Time
	var threadSlug sql.NullString
	parentId := sql.NullInt64{}

	queryJoin := ""
	queryJoinOn := ""
	scanArgs := []interface{}{&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited}
	for _, ralatedItem := range related {
		switch ralatedItem {
		case "user":
			user = &models.User{}
			queryJoin += `, u.name, u.nick, u.email, u.about `
			queryJoinOn += ` JOIN users u ON p.author_nick = u.nick `
			scanArgs = append(scanArgs, &user.Name, &user.Nick, &user.Email, &user.About)
		case "thread":
			thread = &models.Thread{}
			queryJoin += `, t.id, t.slug, t.title, t.author_nick, t.forum_slug, t.message, t.votes, t.created`
			queryJoinOn += ` JOIN threads t ON p.thread_id = t.id`
			scanArgs = append(scanArgs, &thread.Id, &threadSlug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &threadCreated)
		case "forum":
			forum = &models.Forum{}
			queryJoin += `, f.slug, f.title, f.posts, f.threads, f.author_nick`
			queryJoinOn += ` JOIN forums f ON p.forum_slug = f.slug`
			scanArgs = append(scanArgs, &forum.Slug, &forum.Title, &forum.Posts, &forum.Threads, &forum.UserNick)
		}
	}

	err := r.Conn.QueryRow(`SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited`+queryJoin+` FROM posts p`+queryJoinOn+` WHERE p.id=$1`, id).Scan(scanArgs...)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	post.Created = strfmt.DateTime(created.UTC()).String()
	post.ParentId = int(parentId.Int64)

	if thread != nil {
		thread.Created = strfmt.DateTime(threadCreated.UTC()).String()
		thread.Slug = threadSlug.String
	}

	return post, user, forum, thread, nil
}

func (r *Repo) UpdatePost(post *models.Post) (*models.Post, error) {
	var created time.Time
	parentId := sql.NullInt64{}

	// err := r.Conn.QueryRow(`UPDATE posts SET message=COALESCE(NULLIF($1, ''), message), is_edited=true WHERE id=$2 RETURNING id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited`, post.Message, post.Id).Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
	err := r.Conn.QueryRow(`EXECUTE update_post($1, $2)`, post.Message, post.Id).Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
	// fmt.Println(err.Error())
	post.Created = strfmt.DateTime(created.UTC()).String()
	post.ParentId = int(parentId.Int64)
	if err != nil {
		return nil, err
	}
	return post, nil
}
