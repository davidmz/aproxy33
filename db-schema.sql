CREATE TABLE apps
(
  id serial NOT NULL,
  key text NOT NULL,
  secret text NOT NULL,
  user_id integer NOT NULL,
  date timestamp with time zone NOT NULL DEFAULT now(),
  title text NOT NULL,
  description text NOT NULL,
  domains text[] NOT NULL,
  CONSTRAINT apps_pkey PRIMARY KEY (id),
  CONSTRAINT apps_user_id_fkey FOREIGN KEY (user_id)
      REFERENCES users (id) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX apps_key_idx ON apps;

CREATE TABLE atokens
(
  id serial NOT NULL,
  date timestamp with time zone NOT NULL DEFAULT now(),
  app_id integer NOT NULL,
  user_id integer NOT NULL,
  token text NOT NULL,
  perms text[] NOT NULL,
  CONSTRAINT atokens_pkey PRIMARY KEY (id),
  CONSTRAINT atokens_app_id_fkey FOREIGN KEY (app_id)
      REFERENCES apps (id) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT atokens_user_id_fkey FOREIGN KEY (user_id)
      REFERENCES users (id) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX atokens_token_idx ON atokens;

CREATE TABLE users
(
  id serial NOT NULL,
  username text NOT NULL,
  ff_token text NOT NULL,
  date timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT users_pkey PRIMARY KEY (id),
  CONSTRAINT users_username_key UNIQUE (username)
);

CREATE INDEX users_username_idx ON users;