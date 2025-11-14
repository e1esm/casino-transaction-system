-- +goose Up

drop index if exists idx_user_uuid;
create index idx_user_id on transactions using hash(user_id);

-- +goose Down

drop index if exists idx_user_id;
create index idx_user_uuid on transactions(user_id);