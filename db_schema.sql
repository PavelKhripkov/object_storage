create table chunk
(
    id             TEXT    not null
        constraint chunk_pk
            primary key,
    item_id        TEXT    not null,
    position       INTEGER not null,
    file_server_id TEXT
        constraint chunk_file_server_fk
            references file_server,
    file_path      TEXT    not null,
    size           INTEGER not null,
    created        INTEGER,
    modified       INTEGER
);

------------------------------------------

create table container
(
    id          TEXT not null
        constraint container_pk
            primary key,
    name        TEXT not null,
    description TEXT not null,
    parent_id   TEXT not null
        constraint container_container_id_fk
            references container,
    created     INTEGER,
    modified    INTEGER
);

create unique index container_parent_id_name_uindex
    on container (parent_id, name);

------------------------------------------

create table file_server
(
    id          TEXT              not null
        constraint file_server_pk
            primary key,
    name        TEXT              not null,
    params      TEXT,
    type        integer TEXT      not null,
    created     INTEGER,
    modified    INTEGER,
    status      TEXT,
    total_space INTEGER           not null,
    used_space  INTEGER default 0 not null
);

create unique index file_server_name_uindex
    on file_server (name);

------------------------------------------

create table item
(
    id           TEXT not null
        constraint item_pk
            primary key,
    name         TEXT not null,
    container_id TEXT not null,
    chunk_count  INTEGER,
    status       TEXT,
    size         INTEGER,
    created      INTEGER,
    modified     INTEGER
);