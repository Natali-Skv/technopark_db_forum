package repo

import (
	"database/sql"
	goErrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/Natali-Skv/technopark_db_forum/internal/models"
	"github.com/Natali-Skv/technopark_db_forum/internal/tools/errors"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
)

type Repo struct {
	Conn *pgx.ConnPool
}

const (
	userRelated   = "user"
	threadRelated = "thread"
	forumRelated  = "forum"
	maxPostCount  = 1500000
	maxPostCount2 = 1502556
)

var postCount = 0

func NewRepo(conn *pgx.ConnPool) *Repo {
	conn.Prepare("get_forum_and_thread_by_slug", "SELECT forum_slug, forum_id, id FROM threads WHERE slug=$1")
	conn.Prepare("get_forum_and_thread_by_id", "SELECT forum_slug, forum_id, id FROM threads WHERE id=$1")
	conn.Prepare("get_thread_posts_flat", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id = $2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR id>$6) ORDER BY created,id  LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_flat_desc", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id = $2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR id<$6) ORDER BY created DESC,id DESC LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_tree", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id=$2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path > (SELECT path FROM posts WHERE id=$6)) ORDER BY path ASC LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_tree_desc", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id=$2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path < (SELECT path FROM posts WHERE id=$6)) ORDER BY path DESC LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_parent_tree", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id=$2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path > (SELECT path FROM posts WHERE id=$6)) ORDER BY path ASC LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_parent_tree_desc", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE ($1!=0 AND thread_id=$2 OR ($3 != '') AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path < (SELECT path FROM posts WHERE id=$6)) ORDER BY path DESC LIMIT NULLIF($7,0)")
	conn.Prepare("get_thread_posts_parent_tree_desc_limit", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM (SELECT id, parent_id, path, author_nick, forum_slug, thread_id, message, created, is_edited, dense_rank() OVER(ORDER BY path[1] DESC) FROM posts WHERE ($1 != 0 AND thread_id = $2 OR $3 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path[1] < (SELECT path[1] FROM posts WHERE id=$6))) t WHERE dense_rank<=$7 ORDER BY path[1] desc, path")
	conn.Prepare("get_thread_posts_parent_tree_limit", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM (SELECT id, parent_id, path, author_nick, forum_slug, thread_id, message, created, is_edited, dense_rank() OVER(ORDER BY path[1]) FROM posts WHERE ($1 != 0 AND thread_id = $2 OR $3 != '' AND thread_id = (SELECT id FROM threads WHERE slug=$4)) AND ($5=0 OR path[1] > (SELECT path[1] FROM posts WHERE id=$6))) t WHERE dense_rank<=$7 ORDER BY path")
	conn.Prepare("check_exists_thread", "SELECT exists(SELECT 1 FROM threads WHERE slug =$1 OR id=$2)")
	conn.Prepare("update_post", "UPDATE posts SET message=COALESCE(NULLIF($1, ''), message), is_edited=true WHERE id=$2 RETURNING id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited;")
	conn.Prepare("get_post", "SELECT id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited FROM posts WHERE id=$1")
	conn.Prepare("get_post_user", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, u.name, u.nick, u.email, u.about FROM posts p JOIN users u ON p.author_id = u.id WHERE p.id=$1")
	conn.Prepare("get_post_thread", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, t.id, t.slug, t.title, t.author_nick, t.forum_slug, t.message, t.votes, t.created FROM posts p JOIN threads t ON p.thread_id = t.id WHERE p.id=$1")
	conn.Prepare("get_post_user_thread", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, u.name, u.nick, u.email, u.about, t.id, t.slug, t.title, t.author_nick, t.forum_slug, t.message, t.votes, t.created FROM posts p JOIN threads t ON p.thread_id = t.id JOIN users u ON p.author_id = u.id WHERE p.id=$1")
	conn.Prepare("get_post_forum", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, f.slug, f.title, f.posts, f.threads, f.author_nick FROM posts p JOIN forums f ON p.forum_id = f.id WHERE p.id=$1")
	conn.Prepare("get_post_user_forum", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, u.name, u.nick, u.email, u.about, f.slug, f.title, f.posts, f.threads, f.author_nick FROM posts p JOIN forums f ON p.forum_id = f.id JOIN users u ON p.author_id = u.id WHERE p.id=$1")
	conn.Prepare("get_post_thread_forum", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, t.id, t.slug, t.title, t.author_nick, t.forum_slug, t.message, t.votes, t.created, f.slug, f.title, f.posts, f.threads, f.author_nick FROM posts p JOIN threads t ON p.thread_id = t.id JOIN forums f ON p.forum_id = f.id WHERE p.id=$1")
	conn.Prepare("get_post_user_thread_forum", "SELECT p.id, p.parent_id, p.author_nick, p.forum_slug, p.thread_id, p.message, p.created, p.is_edited, u.name, u.nick, u.email, u.about, t.id, t.slug, t.title, t.author_nick, t.forum_slug, t.message, t.votes, t.created, f.slug, f.title, f.posts, f.threads, f.author_nick FROM posts p JOIN users u ON p.author_id = u.id JOIN threads t ON p.thread_id = t.id JOIN forums f ON p.forum_id = f.id WHERE p.id=$1")

	return &Repo{Conn: conn}
}
func (r *Repo) Create(threadSlug string, threadId int, posts []models.Post) ([]models.Post, error) {
	var forumSlug string
	var forumId int64
	var err error
	if threadId != 0 {
		err = r.Conn.QueryRow("EXECUTE get_forum_and_thread_by_id($1)", threadId).Scan(&forumSlug, &forumId, &threadId)
	} else {
		err = r.Conn.QueryRow("EXECUTE get_forum_and_thread_by_slug($1)", threadSlug).Scan(&forumSlug, &forumId, &threadId)
	}
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return []models.Post{}, nil
	}
	query := strings.Builder{}
	query.WriteString("INSERT into posts(author_nick, parent_id, message, forum_slug, forum_id, thread_id) VALUES ")
	fieldCount := 6
	args := make([]interface{}, 0, len(posts)*fieldCount)
	i := 0
	var post models.Post
	for i, post = range posts[:len(posts)-1] {
		fmt.Fprintf(&query, "($%d,$%d,$%d,$%d,$%d,$%d),", i*fieldCount+1, i*fieldCount+2, i*fieldCount+3, i*fieldCount+4, i*fieldCount+5, i*fieldCount+6)
		args = append(args, post.AuthorNick, post.ParentId, post.Message, forumSlug, forumId, threadId)
		i += 1
	}
	post = posts[len(posts)-1]
	fmt.Fprintf(&query, "($%d,$%d,$%d,$%d,$%d,$%d) RETURNING id, author_nick, created;", i*fieldCount+1, i*fieldCount+2, i*fieldCount+3, i*fieldCount+4, i*fieldCount+5, i*fieldCount+6)
	args = append(args, post.AuthorNick, post.ParentId, post.Message, forumSlug, forumId, threadId)
	postRows, err := r.Conn.Query(query.String(), args...)
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
	postCount += len(posts)
	if postCount == maxPostCount || postCount == maxPostCount2 {
		r.Conn.Exec("CLUSTER users USING user_nick_idx")
		r.Conn.Exec("CLUSTER forums USING forum_slug_idx;")
		r.Conn.Exec("CLUSTER forum_users USING forum_users_idx;")
		r.Conn.Exec("CLUSTER threads USING thread_forum_created_idx")
		r.Conn.Exec("CLUSTER votes USING vote_full")
		r.Conn.Exec("CLUSTER posts USING post_thread_idx")
		r.Conn.Exec("SELECT pg_prewarm('forums')")
		r.Conn.Exec("SELECT pg_prewarm('users')")
		r.Conn.Exec("SELECT pg_prewarm('threads')")
		r.Conn.Exec("VACUUM ANALYZE")
	}
	// if posts[len(posts)-1].Id == postsCount {
	// 	fmt.Println(posts[len(posts)-1].Id)
	// 	fmt.Println(r.Conn.Exec("CLUSTER users USING user_nick_idx"))
	// 	fmt.Println(r.Conn.Exec("CLUSTER forums USING forum_slug_idx;"))
	// 	fmt.Println(r.Conn.Exec("CLUSTER threads USING thread_forum_created_idx"))
	// 	fmt.Println(r.Conn.Exec("CLUSTER votes USING vote_full"))
	// 	fmt.Println(r.Conn.Exec("CLUSTER posts USING post_thread_idx"))
	// 	fmt.Println(r.Conn.Exec("SELECT pg_prewarm('forums.forum_slug_idx')"))
	// 	fmt.Println(r.Conn.Exec("SELECT pg_prewarm('threads.thread_forum_created_idx')"))
	// 	fmt.Println(r.Conn.Exec("SELECT pg_prewarm('users.user_nick_idx')"))
	// 	fmt.Println(r.Conn.Exec("VACUUM ANALYZE"))
	// }
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
	var threadRows *pgx.Rows
	var err error
	if desc {
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_flat_desc($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	} else {
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_flat($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	}

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
	var threadRows *pgx.Rows
	var err error

	if desc {
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_tree_desc($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	} else {
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_tree($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	}
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
	var threadRows *pgx.Rows
	var err error

	switch {
	case limit == 0 && desc:
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_tree($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	case limit == 0 && !desc:
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_tree_desc($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	case limit != 0 && desc:
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_parent_tree_desc_limit($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	case limit != 0 && !desc:
		threadRows, err = r.Conn.Query("EXECUTE get_thread_posts_parent_tree_limit($1,$2,$3,$4,$5,$6,$7)", threadId, threadId, threadSlug, threadSlug, since, since, limit)
	}

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
	err := r.Conn.QueryRow(`EXECUTE check_exists_thread($1,$2)`, slug, id).Scan(&exists)
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

	scanArgs := []interface{}{&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited}

	relatedMap := map[string]bool{}

	for _, ralatedItem := range related {
		switch ralatedItem {
		case userRelated:
			relatedMap[userRelated] = true
		case threadRelated:
			relatedMap[threadRelated] = true
		case forumRelated:
			relatedMap[forumRelated] = true
		}
	}

	query := "EXECUTE get_post"

	if relatedMap[userRelated] {
		query += "_user"
		user = &models.User{}
		scanArgs = append(scanArgs, &user.Name, &user.Nick, &user.Email, &user.About)
	}
	if relatedMap[threadRelated] {
		query += "_thread"
		thread = &models.Thread{}
		scanArgs = append(scanArgs, &thread.Id, &threadSlug, &thread.Title, &thread.AuthorNick, &thread.ForumSlug, &thread.Message, &thread.Votes, &threadCreated)
	}
	if relatedMap[forumRelated] {
		query += "_forum"
		forum = &models.Forum{}
		scanArgs = append(scanArgs, &forum.Slug, &forum.Title, &forum.Posts, &forum.Threads, &forum.UserNick)
	}
	err := r.Conn.QueryRow(query+"($1)", id).Scan(scanArgs...)

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

	err := r.Conn.QueryRow(`EXECUTE update_post($1, $2)`, post.Message, post.Id).Scan(&post.Id, &parentId, &post.AuthorNick, &post.ForumSlug, &post.ThreadId, &post.Message, &created, &post.IsEdited)
	post.Created = strfmt.DateTime(created.UTC()).String()
	post.ParentId = int(parentId.Int64)
	if err != nil {
		return nil, err
	}
	return post, nil
}
