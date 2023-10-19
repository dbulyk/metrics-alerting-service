create table if not exists metrics
(
    id    text primary key,
    mtype text not null,
    delta bigint,
    value double precision,
    hash  text
);