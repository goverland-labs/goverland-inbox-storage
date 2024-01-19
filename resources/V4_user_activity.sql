create table user_activity
(
    id          bigserial primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    user_id     uuid not null,
    finished_at timestamp with time zone
);

create index if not exists idx_user_activity_deleted_at on user_activity (deleted_at);
create index if not exists idx_user_activity_user on user_activity (user_id);
