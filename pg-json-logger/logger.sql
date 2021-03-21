/*
pg_json_logger

An auditing logger
Rory Campbell-Lange 19 January 2021

This json logger is a very simple version of the more comprehensive trigger
mechanisms using hstore by 2nd Quadrant at
https://github.com/2ndQuadrant/audit-trigger/, specifically
https://github.com/2ndQuadrant/audit-trigger/blob/master/audit.sql

The logging system below logs the changes due to insert, update and delete
actions on tables which have an appropriate "on each row" trigger. The trigger
registration allows certain columns to be excluded from the log.

The jsonb differencing function is by Dmitry Savinkov on Stack Overflow,
retrieved from https://stackoverflow.com/a/36043269

The main fn_logger function is a simplified version of the 2ndQuadrant hstore
logger suitable for a simple use case and jsonb storage rather than hstore. It
also provides a function for loading the logger and viewing tables with the
logging trigger attached.

Note that the log table described here stores both the table name and
primary key of each table in columns named "t_name" and "n_id"
respectively to allow convenient lookups at the cost of additional
storage. **The primary key of the logged table is assumed to be named "id"**
*/

CREATE TYPE e_log_action_type AS ENUM (
    'insert'
    ,'update'
    ,'delete'
    ,'truncate'
);

CREATE TABLE IF NOT EXISTS log (
    id SERIAL PRIMARY KEY
    ,dt_modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp
    ,e_action e_log_action_type NOT NULL
    ,t_table TEXT NOT NULL
    ,n_id INTEGER
    ,j_changes JSONB
    -- ,j_data JSONB (could be used for verbatim new values)
);

CREATE INDEX IF NOT EXISTS idx_log_table_id ON log (t_table, n_id);

/*
The jsonb differencing function is by Dmitry Savinkov on Stack Overflow.
at https://stackoverflow.com/a/36043269, with a few changes and annotations.
*/
CREATE OR REPLACE FUNCTION fn_jsonb_diff_values(
    newval JSONB
    ,oldval JSONB
) RETURNS JSONB AS $$
DECLARE
  v RECORD;
BEGIN
   -- expand oldval to set of key/value pairs
   FOR v IN SELECT * FROM jsonb_each(oldval) LOOP
     IF newval @> jsonb_build_object(v.key, v.value)
        -- delete the newval key/value pair if newval contains the old key/val
        THEN newval = newval - v.key;
     ELSIF newval ? v.key
        -- ignore key if it just exists in oldval
        THEN CONTINUE;
     ELSE
        -- concatenate the key with a null value to newval
        newval = newval || jsonb_build_object(v.key, 'null');
     END IF;
   END LOOP;
   RETURN newval;
END;
$$ LANGUAGE plpgsql;

/*
The logging function. Note that arguments can be provided to this function
at trigger registration, although these need to be received through TG_ARGV
*/
CREATE OR REPLACE FUNCTION fn_logger() RETURNS TRIGGER AS
$$

DECLARE
    audit_row log%ROWTYPE;
    here_action e_log_action_type;
    -- exclude certain columns from being logged (from 2nd quadrant logger)
    excluded_cols text[] = ARRAY[]::text[];
    j_old jsonb;
    j_new jsonb;
    j_diff jsonb;

BEGIN
    IF TG_WHEN <> 'AFTER' THEN
        RAISE EXCEPTION 'fn_logger may only run as an AFTER trigger';
    END IF;

    IF TG_LEVEL <> 'ROW' THEN
        RAISE EXCEPTION 'fn_logger may only run for ROW triggers';
    END IF;

    /*
    triggers receive arguments in the TG_ARGV array for json dictionary key removal
    see "-" operator https://www.postgresql.org/docs/9.6/functions-json.html; note
    that "-" can be either a text string or text array
    https://stackoverflow.com/a/55654970
    */
    IF TG_NARGS > 0 THEN
        excluded_cols = TG_ARGV;
    END IF;
    j_old := to_jsonb(OLD) - excluded_cols;
    j_new := to_jsonb(NEW) - excluded_cols;


    IF TG_OP = 'UPDATE' THEN
        -- only calculate diff if in UPDATE mode as expensive
        -- return early if no changes
        j_diff := fn_jsonb_diff_values(j_new, j_old);
        IF j_diff = '{}'::jsonb THEN
            RETURN NULL;
        END IF;
    END IF;

    here_action = lower(TG_OP)::e_log_action_type;

    audit_row = ROW(
        nextval('log_id_seq')  -- event_id
        ,current_timestamp     -- action_timestamp
        ,here_action           -- action
        ,TG_TABLE_NAME::text   -- table_name
        ,OLD.id                -- table id
        ,NULL                  -- changes
    );

    IF TG_OP = 'UPDATE' THEN
        audit_row.j_changes = j_diff;

    ELSIF TG_OP = 'DELETE' THEN
        audit_row.j_changes = j_old;

    ELSIF TG_OP = 'INSERT' THEN
        audit_row.n_id = NEW.id;
        audit_row.j_changes = j_new;

    ELSE
        RAISE EXCEPTION 'fn_logger trigger func added as trigger for unhandled case: %', TG_OP;
        RETURN NULL;

    END IF;

    INSERT INTO log VALUES (audit_row.*);

    RETURN NULL;

END;
$$ LANGUAGE plpgsql;
