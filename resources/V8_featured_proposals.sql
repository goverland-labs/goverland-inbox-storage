create table featured_proposals
(
    proposal_id text not null,
    created_at  timestamp with time zone,
    start_at    timestamp with time zone,
    end_at      timestamp with time zone
);

alter table featured_proposals
    add primary key (proposal_id);
