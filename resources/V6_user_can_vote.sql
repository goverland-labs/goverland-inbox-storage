create table user_can_vote
(
    user_id     uuid not null,
    proposal_id text not null,
    created_at  timestamp with time zone
);

alter table user_can_vote
    add primary key (user_id, proposal_id);

create index user_can_vote_created_at_idx on user_can_vote (created_at);
