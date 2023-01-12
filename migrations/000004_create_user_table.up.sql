create table if not exists users (
    id bigserial Primary key,
    create_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now(),
    name TEXT NOT NULL,
    email citext UNIQUE not NULL,
    password_hash bytea NOT NULL,
    activated bool not null,
    version integer not null default 1
);