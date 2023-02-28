create table user_balances (
    user_id bigint primary key,
    balance bigint not null,
    check ( balance >= 0 )
);
