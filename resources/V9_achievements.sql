create table achievements
(
    id          text                    not null
        constraint achievements_pk
            primary key,
    created_at  timestamp default now(),
    deleted_at  timestamp default null,
    title       text                    not null,
    subtitle    text                    not null,
    description text                    not null,
    sort_order  text                    not null,
    exclusive   boolean   default false not null,
    blocked_by  jsonb     default null,
    params      jsonb     default '{}'  not null,
    image_path  text                    not null,
    type        text                    not null
);

create table user_achievements
(
    user_id        uuid,
    created_at     timestamp default now(),
    updated_at     timestamp default now(),
    achieved_at    timestamp default null,
    viewed_at      timestamp default null,
    achievement_id text,
    progress       int       default 0
);

alter table user_sessions
    add app_version text;

alter table user_sessions
    add app_platform text;

alter table user_achievements
    add constraint idx_user_achievements_user_achievement
        unique (user_id, achievement_id);

alter table achievements
    add achievement_message text;

alter table achievements
    alter column achievement_message set default '';

insert into user_achievements (user_id, achievement_id)
    (select u.id, a.id
     from achievements a
              cross join users u
     group by u.id, a.id)
on conflict (user_id, achievement_id) DO NOTHING;

update user_achievements
set achieved_at = now(),
    progress    = 1
where achievement_id = 'early-tester';

alter table achievements
    alter column blocked_by type text using blocked_by::text;

alter table achievements
    rename column image_path to images;

update achievements
set images = '[]'
where images <> '';

alter table achievements
    alter column images type jsonb using images::jsonb;

alter table achievements
    alter column images set default '[]';
