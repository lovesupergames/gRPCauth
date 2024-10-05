create table if not exists users
(
    id integer primary key,
    email text unique not null,
    pass_hash BLOB not null
);
create index if not exists idx_email on users(email);

create table if not exists apps
(
    id integer primary key ,
    name text not null unique,
    secret text not null unique
);