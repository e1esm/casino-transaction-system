-- +goose Up

create extension if not exists pgcrypto;

alter table transactions add column t_hash text unique;

update transactions
set t_hash = encode(
        digest(
                user_id::text ||
                transaction_type ||
                amount::text ||
                to_char(transaction_time at time zone 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
                'sha256'
        ),
        'hex'
                  )
where t_hash is null;
-- +goose Down

alter table transactions drop column t_hash;