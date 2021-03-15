create table if not exists contacts (
    id serial primary key,
    url text not null,
    method text not null,
    protocol text not null,
    headers text not null,
    body text default null
);
