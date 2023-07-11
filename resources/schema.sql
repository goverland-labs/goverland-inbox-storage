create table users
(
    id          text not null
        primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    device_uuid text
);

create index if not exists idx_users_deleted_at
    on users (deleted_at);

create table user_subscriptions
(
    id          text not null
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id    text,
    dao_id     text
);

create index if not exists idx_subscriptions_deleted_at
    on user_subscriptions (deleted_at);

create table global_subscriptions
(
    id            bigserial
        primary key,
    created_at    timestamp with time zone,
    updated_at    timestamp with time zone,
    deleted_at    timestamp with time zone,
    subscriber_id text,
    dao_id        text
);

create index if not exists idx_global_subscriptions_deleted_at
    on global_subscriptions (deleted_at);
