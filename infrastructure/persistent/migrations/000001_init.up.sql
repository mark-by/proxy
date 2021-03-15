create table if not exists requests (
    id serial primary key,
    raw text not null,
    url text not null
);
