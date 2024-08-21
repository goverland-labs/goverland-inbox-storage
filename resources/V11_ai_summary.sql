create table proposal_ai_summary
(
    proposal_id text                    not null,
    created_at  timestamp default now() not null,
    ai_summary  text                    not null
);

create index proposal_ai_summary__proposal_id_idx
    on proposal_ai_summary (proposal_id);

create table ai_requests
(
    created_at  timestamp default now() not null,
    user_id     uuid                    not null,
    address     text                    not null,
    proposal_id text                    not null
);

create index ai_requests_by_address_index
    on ai_requests (address);

create index ai_requests_by_user_index
    on ai_requests (user_id);
