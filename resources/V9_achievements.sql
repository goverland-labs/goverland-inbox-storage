create table achievements
(
    id          text                  not null
        constraint achievements_pk
            primary key,
    name        text                  not null,
    description text                  not null,
    goals       text                  not null,
    exclusive   boolean default false not null,
    blocked_by  text,
    rules       jsonb   default '[]'  not null,
    image_path  text                  not null
);

create table user_achievements
(
    created_at timestamp default now(),
    user_id uuid,
    achievement_id text,
    progress float,
    last_viewed timestamp default null
);

alter table user_sessions
    add app_version text;
