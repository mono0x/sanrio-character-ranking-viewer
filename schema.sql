-- for PostgreSQL

CREATE TABLE status (
  id BIGINT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  source TEXT NOT NULL,
  PRIMARY KEY (id)
);
CREATE INDEX index_status_created_at_date ON statuses USING btree (
  DATE_PART('year'::text, created_at),
  DATE_PART('month'::text, created_at),
  DATE_PART('day'::text, created_at)
);

CREATE TABLE character (
  id SERIAL NOT NULL,
  name VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);
CREATE UNIQUE INDEX unique_index_character_name ON character USING btree (name);

CREATE TABLE ranking (
  id SERIAL NOT NULL,
  name VARCHAR(255) NOT NULL,
  started_on DATE NOT NULL,
  ended_on DATE NOT NULL,
  PRIMARY KEY (id)
);
CREATE INDEX index_ranking_ended_on_started_on ON ranking USING btree (ended_on, started_on);

CREATE TABLE vote (
  ranking_id INTEGER NOT NULL,
  character_id INTEGER NOT NULL,
  status_id BIGINT NOT NULL,
  PRIMARY KEY (ranking_id, character_id, status_id)
);
CREATE UNIQUE INDEX unique_index_vote_status_id ON vote USING btree (status_id);
CREATE UNIQUE INDEX unique_index_vote_character_id_status_id ON vote USING btree (character_id, status_id);
