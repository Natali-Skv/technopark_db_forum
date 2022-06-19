CREATE EXTENSION IF NOT EXISTS citext;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS votes CASCADE;
CREATE TABLE users
(
    -- id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name text NOT NULL,
    nick citext COLLATE "C"  UNIQUE NOT NULL PRIMARY KEY,
    email citext  UNIQUE NOT NULL,
    about text 
);
CREATE TABLE forums
(
    slug citext UNIQUE NOT NULL PRIMARY KEY,
    title text NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users NOT NULL,
    threads integer DEFAULT 0,
    posts integer DEFAULT 0
);
CREATE TABLE threads 
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug citext UNIQUE,
    title text NOT NULL,
    author_nick citext COLLATE "C" REFERENCES users NOT NULL,
    forum_slug citext REFERENCES forums NOT NULL,
    message text NOT NULL,
    votes integer DEFAULT 0,
    created timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE posts
(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    author_nick citext COLLATE "C" REFERENCES users NOT NULL,
	parent_id BIGINT REFERENCES posts,
    description text,
    message text NOT NULL,
	is_edited boolean DEFAULT false,
    forum_slug citext REFERENCES forums NOT NULL,
    thread_id integer REFERENCES threads NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    path BIGINT[] default array []::INTEGER[]
);
CREATE TABLE votes 
(
    user_nick citext REFERENCES users NOT NULL,
	thread_id BIGINT REFERENCES threads NOT NULL,
    vote integer NOT NULL,
    UNIQUE (user_nick, thread_id)
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
BEGIN
    SELECT nick INTO NEW.author_nick FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
    END IF;
    UPDATE forums SET threads = threads + 1 WHERE slug = NEW.forum_slug RETURNING slug INTO NEW.forum_slug;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA3';
    END IF;
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
    SELECT nick INTO NEW.author_nick FROM users WHERE nick=NEW.author_nick;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'AAAA1';
        RETURN NULL;
    END IF;
    UPDATE forums SET posts = posts + 1 WHERE slug = NEW.forum_slug;
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
        -- UPDATE threads SET branches = branches + 1 WHERE id = NEW.thread_id RETURNING branches INTO branches_count;
        NEW.path = ARRAY[NEW.id];
    END IF;
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
------
CREATE TRIGGER insert_vote_to_thread_tg AFTER INSERT ON votes
FOR EACH ROW EXECUTE FUNCTION insert_vote_to_thread();

CREATE TRIGGER update_vote_to_thread_tg AFTER UPDATE ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_to_thread();

CREATE TRIGGER get_user_nick_tg BEFORE INSERT ON votes
FOR EACH ROW EXECUTE FUNCTION get_user_nick();
------
CREATE TRIGGER threads_tg BEFORE INSERT ON threads
FOR EACH ROW EXECUTE FUNCTION insert_threads_tg();
------
CREATE TRIGGER posts_tg BEFORE INSERT ON posts
FOR EACH ROW EXECUTE FUNCTION insert_posts_tg();

CREATE TRIGGER update_posts_tg BEFORE UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION update_posts_tg();






-- SELECT name,nick,email,about FROM 
--     (SELECT author_nick FROM threads WHERE forum_slug='5LcyNwg0_L5Ur' AND lower(author_nick) > lower('gCIXNSQGY9Itp.bill') UNION DISTINCT 
--     SELECT author_nick FROM posts WHERE forum_slug='5LcyNwg0_L5Ur' AND lower(author_nick) > lower('gCIXNSQGY9Itp.bill')) t JOIN users u ON t.author_nick=u.nick  ORDER BY lower(nick) LIMIT 4
-- 5LcyNwg0_L5Ur gCIXNSQGY9Itp.Zod 4