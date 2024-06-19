create table user_settings
(
    user_id    uuid                    not null,
    type       text                    not null,
    value      jsonb     default '{}',
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    deleted_at timestamp default null
);

comment on column user_settings.type is 'setting type: push-details, etc';

create index user_settings_user_id_idx
    on user_settings (user_id);

