DROP TABLE IF EXISTS emails CASCADE;

CREATE
    EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE
    EXTENSION IF NOT EXISTS CITEXT;
-- CREATE EXTENSION IF NOT EXISTS postgis;
-- CREATE EXTENSION IF NOT EXISTS postgis_topology;


CREATE TABLE emails
(
    email_id     UUID PRIMARY KEY                  DEFAULT uuid_generate_v4(),
    "to"         VARCHAR(500)             NOT NULL,
    "from"       VARCHAR(250)             NOT NULL,
    subject      VARCHAR(250)             NOT NULL,
    body         VARCHAR(5000)            NOT NULL,
    content_type VARCHAR(250)             NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
)
