create table auth_nonces
(
    address    text,
    nonce      text,
    expired_at timestamp with time zone
);

alter table auth_nonces
    add primary key (address, nonce, expired_at);
