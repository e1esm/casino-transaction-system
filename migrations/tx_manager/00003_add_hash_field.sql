-- +goose Up

alter table transactions add column t_hash text unique;

UPDATE transactions
SET t_hash = encode(
        digest(
                user_id::text ||
                transaction_type ||
                amount::text ||
                to_char(transaction_time AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
                'sha256'
        ),
        'hex'
                  )
where t_hash is null;
-- +goose Down

alter table transactions drop column t_hash;