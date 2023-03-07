create table operations (
    id bigserial primary key,
    amount bigint not null,
    created_at timestamptz not null,
    description text,
    sender_id bigint,
    recipient_id bigint
);
