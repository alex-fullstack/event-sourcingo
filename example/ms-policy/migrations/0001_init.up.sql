CREATE SCHEMA IF NOT EXISTS "es";

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS es.aggregates (
    id UUID PRIMARY KEY,
    version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS es.events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    command_type INTEGER NOT NULL,
    version INTEGER NOT NULL,
    event_type INTEGER NOT NULL,
    payload JSON NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS es.transactions (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    sequence_id XID8 DEFAULT pg_current_xact_id() NOT NULL
);

CREATE TABLE IF NOT EXISTS es.subscription (
    id INTEGER PRIMARY KEY,
    last_sequence_id XID8 NOT NULL
);

ALTER TABLE es.events ADD CONSTRAINT  events_aggregates_id_fk FOREIGN KEY (aggregate_id) REFERENCES es.aggregates (id) DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE es.events ADD CONSTRAINT  events_transaction_id_fk FOREIGN KEY (transaction_id) REFERENCES es.transactions (id) DEFERRABLE INITIALLY DEFERRED;
CREATE UNIQUE INDEX aggregate_id_version_idx ON es.events (aggregate_id, version);

INSERT INTO es.subscription (id, last_sequence_id) VALUES (1, '0'::xid8) ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION es.notify_transactions() RETURNS TRIGGER AS
    $$
    DECLARE
        channel TEXT := 'es.transaction-handled';
    BEGIN
        PERFORM pg_notify(channel, row_to_json(NEW)::text);
        RETURN NEW;
    END;
    $$
    LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER notify_transactions_trigger
  AFTER INSERT ON es.transactions
  FOR EACH ROW
  EXECUTE PROCEDURE es.notify_transactions();