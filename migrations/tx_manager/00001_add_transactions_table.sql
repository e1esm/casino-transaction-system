-- +goose Up
create table transactions(
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    transaction_type varchar(10) not null,
    amount int not null,
    transaction_time timestamptz not null
);

create index idx_user_uuid on transactions(user_id);
create index idx_transaction_type on transactions(transaction_type);

-- +goose Down

drop index idx_user_uuid;
drop index idx_transaction_type;
drop table transactions;