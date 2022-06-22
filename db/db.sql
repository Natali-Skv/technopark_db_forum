
CREATE EXTENSION IF NOT EXISTS citext;
DROP TABLE IF EXISTS forum_users CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS votes CASCADE;
DEALLOCATE ALL;

CREATE EXTENSION IF NOT EXISTS citext;

ALTER SYSTEM SET
    checkpoint_completion_target = '0.9';
ALTER SYSTEM SET
    wal_buffers = '6912kB';
ALTER SYSTEM SET
    default_statistics_target = '100';
ALTER SYSTEM SET
    random_page_cost = '1.1';
ALTER SYSTEM SET
    effective_io_concurrency = '200';


CREATE UNLOGGED TABLE users
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name text NOT NULL,
    nick citext COLLATE "C"  UNIQUE NOT NULL,
    email citext  UNIQUE NOT NULL,
    about text 
);

CREATE  INDEX idx_users_id ON users(id);
CREATE UNIQUE INDEX idx_users_nick ON users(nick);
CLUSTER users USING idx_users_nick;

CREATE UNLOGGED TABLE forums
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug citext UNIQUE NOT NULL,
    title text NOT NULL,
    -- author_id BIGINT REFERENCES users NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
    threads integer DEFAULT 0,
    posts integer DEFAULT 0
);

CREATE UNIQUE INDEX idx_forums_id ON forums(id);
CREATE UNIQUE INDEX idx_forums_slug_id ON forums(slug, id);
CREATE UNIQUE INDEX idx_forums_id_slug ON forums(id, slug);
CLUSTER forums USING idx_forums_id;

CREATE UNLOGGED TABLE threads 
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug citext UNIQUE,
    title text NOT NULL,
    -- author_id BIGINT REFERENCES users NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
    forum_id BIGINT REFERENCES forums NOT NULL,
    forum_slug citext REFERENCES forums(slug) NOT NULL,
    message text NOT NULL,
    votes integer DEFAULT 0,
    created timestamp with time zone DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX idx_thread_id ON threads(slug);
CREATE INDEX idx_threads_created ON threads(created);
CREATE INDEX idx_threads_slug_id ON threads(slug, id);
CREATE INDEX idx_threads_id_slug ON threads(id, slug);
CREATE INDEX idx_threads_forum_created ON threads(forum_id, created);
CLUSTER threads USING idx_threads_forum_created;

CREATE UNLOGGED TABLE posts
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    author_id BIGINT REFERENCES users NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
	parent_id BIGINT REFERENCES posts,
    description text,
    message text NOT NULL,
	is_edited boolean DEFAULT false,
    forum_slug citext REFERENCES forums(slug) NOT NULL,
    forum_id BIGINT REFERENCES forums NOT NULL,
    thread_id integer REFERENCES threads NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    path BIGINT[] default array []::INTEGER[]
);

CREATE INDEX idx_post_threadID_created_id ON posts(thread_id, created, id);
CREATE INDEX idx_post_threadID_path ON posts(thread_id, path);
CREATE INDEX idx_posts_threadID_root_path ON posts (thread_id, (path[1]), path);
CREATE INDEX idx_post_threadID_id_parentNull_id ON posts(thread_id, id) WHERE parent_id IS NULL;
CREATE INDEX idx_posts_id ON posts (id);
CREATE INDEX idx_posts_id_full ON posts (id, parent_id, thread_id , message, is_edited, created, forum_slug, author_nick) ;
CREATE INDEX idx_post_threadID_ID_parentID ON posts(thread_id, id, parent_id);

CREATE UNLOGGED TABLE votes 
(
    -- author_id BIGINT REFERENCES users NOT NULL,
    user_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
	thread_id BIGINT REFERENCES threads NOT NULL,
    vote integer NOT NULL,
    UNIQUE (user_nick, thread_id)
);
ALTER TABLE ONLY votes ADD CONSTRAINT votes_user_thread_unique UNIQUE (user_nick, thread_id);
CLUSTER votes USING votes_user_thread_unique;CLUSTER votes USING votes_user_thread_unique;

CREATE UNLOGGED TABLE forum_users 
(
    user_id BIGINT REFERENCES users NOT NULL,
    -- user_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
	-- forum_slug citext REFERENCES forums NOT NULL,
    forum_id BIGINT REFERENCES forums NOT NULL,
    UNIQUE (user_id, forum_id)
);

CREATE UNIQUE INDEX idx_forum_users_slug ON forum_users(forum_id, user_id );
CLUSTER forum_users USING idx_forum_users_slug;


---------------------------FUNCTIONS----------------------------
CREATE OR REPLACE FUNCTION get_author_nick() RETURNS TRIGGER AS
$$
BEGIN
    -- SELECT nick,id INTO NEW.author_nick, NEW.author_id FROM users WHERE nick=NEW.author_nick;
    SELECT nick INTO NEW.author_nick FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
    END IF;
   RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_user_nick() RETURNS TRIGGER AS
$$
BEGIN
    RAISE NOTICE 'get_user_nick %', NEW;
    SELECT nick INTO NEW.user_nick FROM users WHERE nick=NEW.user_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
    END IF;
   RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION insert_vote_to_thread() RETURNS TRIGGER AS
$$
BEGIN
    RAISE NOTICE 'insert_vote_to_thread %', NEW;
    UPDATE threads SET votes = votes + NEW.vote WHERE id = NEW.thread_id;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_vote_to_thread() RETURNS TRIGGER AS
$$
BEGIN
    RAISE NOTICE 'update_vote_to_thread %', NEW;
    UPDATE threads SET votes = votes - OLD.vote + NEW.vote WHERE id = NEW.thread_id;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;


------главный триггер вставки треда
CREATE OR REPLACE FUNCTION insert_threads_tg() RETURNS TRIGGER AS
$$
DECLARE
    author_id bigint;
BEGIN
    SELECT nick,id INTO NEW.author_nick,author_id FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
    END IF;
    UPDATE forums SET threads = threads + 1 WHERE slug = NEW.forum_slug RETURNING slug,id INTO NEW.forum_slug,NEW.forum_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA3';
    END IF;
    INSERT INTO forum_users(user_id,forum_id) VALUES(author_id,NEW.forum_id) ON CONFLICT DO NOTHING;
   RETURN NEW;
END
$$ LANGUAGE plpgsql;

------ главный триггер обновления постов
CREATE OR REPLACE FUNCTION insert_posts_tg() RETURNS TRIGGER AS
$$
DECLARE
parent_path bigint[];
correct_parent boolean;
BEGIN
    SELECT nick,id INTO NEW.author_nick,NEW.author_id FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
        RETURN NULL;
    END IF;
    UPDATE forums SET posts = posts + 1 WHERE id = NEW.forum_id;
    IF (NEW.parent_id IS NOT NULL) AND (NEW.parent_id != 0) THEN 
        SELECT thread_id=NEW.thread_id INTO correct_parent FROM posts WHERE id = NEW.parent_id;
        IF NOT FOUND OR NOT correct_parent  THEN
            RAISE EXCEPTION USING ERRCODE = 'AAAA0';
            RETURN NULL;
        END IF; 
        SELECT path FROM posts WHERE id=NEW.parent_id INTO parent_path;
        NEW.path = parent_path || NEW.id;
    ELSE    
        NEW.parent_id=NULL;
        NEW.path = ARRAY[NEW.id];
    END IF;
    INSERT INTO forum_users(user_id,forum_id) VALUES(NEW.author_id,NEW.forum_id) ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_posts_tg() RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.message = NEW.message THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

---------------------------TRIGGERS-----------------------------
CREATE TRIGGER insert_vote_to_thread_tg AFTER INSERT ON votes
FOR EACH ROW EXECUTE FUNCTION insert_vote_to_thread();

CREATE TRIGGER update_vote_to_thread_tg AFTER UPDATE ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_to_thread();

CREATE TRIGGER get_user_nick_tg BEFORE INSERT ON votes
FOR EACH ROW EXECUTE FUNCTION get_user_nick();

CREATE TRIGGER threads_tg BEFORE INSERT ON threads
FOR EACH ROW EXECUTE FUNCTION insert_threads_tg();

CREATE TRIGGER posts_tg BEFORE INSERT ON posts
FOR EACH ROW EXECUTE FUNCTION insert_posts_tg();

CREATE TRIGGER update_posts_tg BEFORE UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION update_posts_tg();

CREATE TRIGGER get_author_nick_tg BEFORE INSERT ON forums
FOR EACH ROW EXECUTE FUNCTION get_author_nick();

---------------------------INDEXES-----------------------------

-- CREATE INDEX post_first_path_thread_idx ON posts ((posts.path[1]), thread_id);
-- CREATE INDEX post_first_parent_id_index ON posts ((posts.path[1]), id);
-- CREATE INDEX post_first_parent_index ON posts ((posts.path[1]));
-- CREATE INDEX post_path_index ON posts ((posts.path));
-- CREATE INDEX post_thread_index ON posts (thread_id); -- -
-- CREATE INDEX post_thread_id_index ON posts (thread_id, id); -- +

-- CREATE INDEX forum_slug_lower_index ON forums (slug); -- +

-- CREATE INDEX users_nickname_lower_index ON users (lower(users.nick));
-- CREATE INDEX users_nickname_index ON users ((users.nick));
-- CREATE INDEX users_email_index ON users (lower(email));

-- CREATE INDEX users_forum_forum_user_index ON forum_users (lower(forum_slug), user_nick);
-- CREATE INDEX users_forum_user_index ON forum_users (user_nick);

-- CREATE INDEX thread_slug_lower_index ON threads (lower(slug));
-- CREATE INDEX thread_slug_index ON threads (slug);
-- CREATE INDEX thread_slug_id_index ON threads (lower(slug), id);
-- CREATE INDEX thread_forum_lower_index ON threads (lower(forum_slug)); -- +
-- CREATE INDEX thread_id_forum_index ON threads (id, forum_slug);
-- CREATE INDEX thread_created_index ON threads (created);

-- CREATE INDEX vote_nickname ON votes (lower(user_nick), thread_id, vote); -- +

-- -- NEW INDEXES
-- CREATE INDEX post_path_id_index ON posts (id, (posts.path));
-- CREATE INDEX post_thread_path_id_index ON posts (thread_id, (posts.parent_id), id);

-- CREATE INDEX users_forum_forum_index ON forum_users (forum_slug); -- +

-- -------
-- drop index if exists post_first_path_thread_idx;
-- drop index if exists post_first_parent_id_index;
-- drop index if exists post_first_parent_index;
-- drop index if exists post_path_index;
-- drop index if exists post_thread_index;
-- drop index if exists post_thread_id_index;
-- drop index if exists forum_slug_lower_index;
-- drop index if exists users_nickname_lower_index;
-- drop index if exists users_nickname_index;
-- drop index if exists users_email_index;
-- drop index if exists users_forum_forum_user_index;
-- drop index if exists users_forum_user_index;
-- drop index if exists thread_slug_lower_index;
-- drop index if exists thread_slug_index;
-- drop index if exists thread_slug_id_index;
-- drop index if exists thread_forum_lower_index;
-- drop index if exists thread_id_forum_index;
-- drop index if exists thread_created_index;
-- drop index if exists vote_nickname;

-- ------------------------------prepare-------------------------
-- PREPARE  update_post(text,bigint) AS
--     UPDATE posts SET message=COALESCE(NULLIF($1, ''), message), is_edited=true WHERE id=$2 RETURNING id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited;
-- -- EXECUTE update_post('message', 1);


-- UPDATE posts SET message=COALESCE(NULLIF($1, ''), message), is_edited=true WHERE id=$2 RETURNING id, parent_id, author_nick, forum_slug, thread_id, message, created, is_edited

-- drop index if exists index_post_by_thread_path;
-- create unique index if not exists index_post_by_thread_path on posts (thread_id, path);

-- drop index if exists index_post_by_thread;
-- create index if not exists index_post_by_thread on posts (thread_id);


-- drop index if exists index_thread_by_slug;

-- drop index if exists index_thread_by_forum;
-- create index if not exists index_thread_by_forum on threads (forum_slug);

-- drop index if exists index_thread_by_created;
-- create index if not exists index_thread_by_created on threads (created);

------------------

-- create index if not exists user_email_idx on users using hash (email);
-- create index if not exists user_nick_idx on users using hash (nick);

-- create index if not exists forum_slug_idx on forums using hash (slug);

-- create index if not exists thread_slug_idx on threads using hash (slug);

-- drop index if exists index_forum_user_by_forum;
-- create index if not exists forum_users_idx on forum_users (forum_id, user_id);
--------------

-- CREATE INDEX IF NOT EXISTS user_nickname_hash ON users using hash (nick); -- common, with hash faster than with default b-tree
-- CREATE INDEX IF NOT EXISTS user_nickname_email ON users (nick, email); -- GetUsersByEmailOrNickname

-- CREATE INDEX IF NOT EXISTS forum_slug_hash ON forums using hash (slug); -- common, with hash faster than with default b-tree

-- CREATE INDEX IF NOT EXISTS thread_slug_hash ON threads using hash (slug); -- GetForumThreadBySlug
-- CREATE INDEX IF NOT EXISTS thread_forum_hash ON threads using hash (forum_slug); -- common
-- CREATE INDEX IF NOT EXISTS thread_forum_created ON threads (forum_slug, created); -- GetForumThreads

-- CREATE INDEX IF NOT EXISTS post_thread_path ON posts (thread_id, path); -- GetPostsFlat (first column), GetPostsTree, GetPostsParentTree
-- CREATE INDEX IF NOT EXISTS post_path_complex ON posts ((path[1]), path); -- GetPostsParentTree, crucial

-- CREATE INDEX IF NOT EXISTS forum_users_forum_hash ON forum_users (forum_id, user_id); -- GetForumUsers

-- -- CREATE UNIQUE INDEX IF NOT EXISTS votes_less ON votes (user_nick, thread_id); -- VoteExists
-- -- CREATE UNIQUE INDEX IF NOT EXISTS votes_more ON votes (nickname, thread, voice); -- UpdateVote


vacuum analyze;