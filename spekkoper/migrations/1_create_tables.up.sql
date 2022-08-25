CREATE TABLE query
(
    id              TEXT NOT NULL,
    query           TEXT NOT NULL,
    category        INT,
    sub_category    INT,
    postcode        TEXT,
    distance_meters int,
    PRIMARY KEY (id)

);

CREATE TABLE query_result
(
    query_id  TEXT NOT NULL,
    result_id TEXT NOT NULL,
    PRIMARY KEY (query_id, result_id),
    CONSTRAINT fk_query
        FOREIGN KEY (query_id)
            REFERENCES query (id)
);