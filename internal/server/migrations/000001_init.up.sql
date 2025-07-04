CREATE TABLE metrics (
     id VARCHAR(255) NOT NULL,
     type VARCHAR(10) NOT NULL CHECK (type IN ('counter', 'gauge')),
     delta BIGINT,
     value DOUBLE PRECISION,
     PRIMARY KEY (id, type)
);
