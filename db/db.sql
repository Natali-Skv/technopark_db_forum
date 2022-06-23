
CREATE EXTENSION pg_prewarm;
CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS forum_users CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS votes CASCADE;
DEALLOCATE ALL;

CREATE UNLOGGED TABLE users
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name text,
    nick citext COLLATE "C"  UNIQUE NOT NULL,
    email citext  UNIQUE NOT NULL,
    about text 
);

CREATE UNLOGGED TABLE forums
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug citext UNIQUE NOT NULL,
    title text,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
    threads integer DEFAULT 0,
    posts integer DEFAULT 0
);

CREATE UNLOGGED TABLE threads 
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug citext UNIQUE,
    title text,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
    forum_id BIGINT REFERENCES forums NOT NULL,
    forum_slug citext REFERENCES forums(slug) NOT NULL,
    message text,
    votes integer DEFAULT 0,
    created timestamp with time zone DEFAULT now()
);

CREATE UNLOGGED TABLE posts
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    author_id BIGINT REFERENCES users NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
	parent_id BIGINT REFERENCES posts,
    description text,
    message text,
	is_edited boolean DEFAULT false,
    forum_slug citext REFERENCES forums(slug) NOT NULL,
    forum_id BIGINT REFERENCES forums NOT NULL,
    thread_id integer REFERENCES threads NOT NULL,
    thread_slug citext,
    created timestamp with time zone DEFAULT now(),
    path BIGINT[] default array []::INTEGER[]
);

CREATE UNLOGGED TABLE votes 
(
    user_nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
	thread_id BIGINT REFERENCES threads NOT NULL,
    vote integer NOT NULL,
    UNIQUE (user_nick, thread_id)
);

CREATE UNLOGGED TABLE forum_users 
(
    nick citext COLLATE "C" REFERENCES users(nick) NOT NULL,
    email citext NOT NULL,
    name text,
    about text,
    forum_slug citext REFERENCES forums(slug) NOT NULL,
    UNIQUE (nick, forum_slug)
);
---------------------------FUNCTIONS----------------------------
CREATE OR REPLACE FUNCTION get_author_nick() RETURNS TRIGGER AS
$$
BEGIN
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
    UPDATE threads SET votes = votes + NEW.vote WHERE id = NEW.thread_id;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_vote_to_thread() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE threads SET votes = votes - OLD.vote + NEW.vote WHERE id = NEW.thread_id;
    RETURN NEW;
END
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION insert_threads_tg() RETURNS TRIGGER AS
$$
DECLARE
author_email text;
author_name text;
author_about text;
BEGIN
    SELECT nick,email,name,about INTO NEW.author_nick,author_email,author_name,author_about FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
    END IF;
    UPDATE forums SET threads = threads + 1 WHERE slug = NEW.forum_slug RETURNING slug,id INTO NEW.forum_slug,NEW.forum_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA3';
    END IF;
    INSERT INTO forum_users(nick,email,name,about,forum_slug) VALUES(NEW.author_nick,author_email,author_name,author_about,NEW.forum_slug) ON CONFLICT DO NOTHING;
   RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION insert_posts_tg() RETURNS TRIGGER AS
$$
DECLARE
parent_path bigint[];
correct_parent boolean;
author_email text;
author_name text;
author_about text;
BEGIN
    SELECT nick,email,name,about,id INTO NEW.author_nick,author_email,author_name,author_about,NEW.author_id FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
        RETURN NULL;
    END IF;
    UPDATE forums SET posts = posts + 1 WHERE id = NEW.forum_id;
    IF NEW.parent_id != 0 THEN 
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
    INSERT INTO forum_users(nick,email,name,about,forum_slug) VALUES(NEW.author_nick,author_email,author_name,author_about,NEW.forum_slug) ON CONFLICT DO NOTHING;
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

CREATE INDEX IF NOT EXISTS user_nick_idx ON users (nick);
CREATE INDEX IF NOT EXISTS user_email_idx ON users USING hash (email); 

CREATE INDEX IF NOT EXISTS forum_slug_idx ON forums (slug);
CREATE INDEX IF NOT EXISTS thread_slug_idx ON threads USING hash(slug); 

CREATE INDEX IF NOT EXISTS thread_forum_slug_idx ON threads (forum_slug); 
CREATE INDEX IF NOT EXISTS thread_forum_created_idx ON threads (forum_slug, created);

CREATE INDEX IF NOT EXISTS post_thread_idx ON posts (thread_id);
CREATE INDEX IF NOT EXISTS post_created_id_idx ON posts (created,id);
CREATE INDEX IF NOT EXISTS post_thread_created_id_idx ON posts (thread_id,created,id);
CREATE INDEX IF NOT EXISTS post_thread_path_idx ON posts (thread_id, path);
CREATE INDEX IF NOT EXISTS post_thread_path1_idx ON posts (thread_id,(path[1]), path);

CREATE INDEX IF NOT EXISTS forum_users_idx ON forum_users (forum_slug, nick);

CREATE UNIQUE INDEX IF NOT EXISTS vote ON votes (user_nick, thread_id);
CREATE UNIQUE INDEX IF NOT EXISTS vote_full ON votes (user_nick, thread_id, vote); 

VACUUM;