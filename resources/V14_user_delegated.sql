create table user_delegated
(
    id         bigserial
        primary key,
    created_at timestamp with time zone not null,
    user_id    uuid                     not null,
    dao_id     text                     not null,
    tx_hash    text                     not null,
    delegates  text                     not null,
    expiration timestamp with time zone
);
