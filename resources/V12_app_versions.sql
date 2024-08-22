create table app_versions
(
    version     text not null,
    platform    text,
    created_at  timestamp default now(),
    description text
);

comment on column app_versions.version is 'semver notation: ex 1.2.3';

comment on column app_versions.platform is 'platform type: iOS, android etc';

comment on column app_versions.description is 'markdown description';

