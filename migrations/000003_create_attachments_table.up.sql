CREATE TABLE attachments (
    id bigserial primary key,
    message_id integer references messages(id) on delete cascade,
    file_path varchar not null,
    file_type varchar not null,
    file_size integer not null,
    created_at TIMESTAMP default CURRENT_TIMESTAMP
);