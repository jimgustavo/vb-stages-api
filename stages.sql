CREATE TABLE stages (
    id SERIAL PRIMARY KEY,
    stage_name VARCHAR(255) NOT NULL,
    stages JSONB NOT NULL
);
