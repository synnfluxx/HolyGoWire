CREATE TABLE messages (
    id bigserial primary key,
    content varchar,
    username varchar,
    send_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);