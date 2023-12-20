alter table users
    add column role text not null default 'GUEST';
alter table users
    add column address text;

create unique index if not exists users_address_idx on users (address) where address is not null and address != '' and deleted_at is null;

create table user_sessions
(
    id          uuid not null primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    user_id     uuid not null,
    device_uuid text,
    device_name text
);

create index if not exists idx_user_sessions_deleted_at on user_sessions (deleted_at);
create index if not exists idx_user_sessions_user_id on user_sessions (user_id);

