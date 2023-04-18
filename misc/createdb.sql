CREATE TABLE application (
    app_key text NOT NULL,
    name text NOT NULL,
    creation_time bigint NOT NULL
);

CREATE TABLE event (
    id integer NOT NULL,
    client_key text NOT NULL,
    app_key text NOT NULL,
    "time" bigint NOT NULL,
    platform text NOT NULL,
    ip text NOT NULL,
    country text,
    version text NOT NULL,
    name text NOT NULL,
    data text
);
