create table recently_viewed
(
    id         bigserial
        primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id    uuid not null,
    type       text,
    type_id    text
);

create index idx_recently_viewed_deleted_at
    on recently_viewed (deleted_at);

create index idx_recently_viewed_user_created_at
    on recently_viewed (user_id, created_at desc, type);
