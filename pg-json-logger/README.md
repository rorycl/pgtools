# pg-json-logger

A PostgreSQL json logger/auditor using triggers

Rory Campbell-Lange  
version 0.2 : 21 January 2021

## Introduction

A trigger-based PostgreSQL json logger/auditor, based closely on the
much more comprehensive hstore-based auditor by 2ndQuadrant at
https://github.com/2ndQuadrant/audit-trigger/, specifically
https://github.com/2ndQuadrant/audit-trigger/blob/master/audit.sql

The logging system logs the changes due to insert, update and delete
actions on tables which have an appropriate "on each row" trigger,
storing the results in a jsonb column in the log table. The trigger
registration allows certain columns to be excluded from the log.
Statement level events are not supported.

The jsonb differencing function is by Dmitry Savinkov on Stack Overflow,
retrieved from https://stackoverflow.com/a/36043269

The main `fn_logger` function is a simplified version of the 2ndQuadrant
hstore logger suitable for a simple use case and jsonb storage rather
than hstore. 2ndQuadrant's solution also provides a function for loading
the logger and viewing tables with the logging trigger attached.

Note that the log table described here stores both the table name and
primary key of each table as separate columns to allow convenient
lookups at the cost of additional storage. Note also that *the primary
key of the logged table is assumed to be named "id"*.

The system can be easily extended to store the entire old and new row
contents on each insert, update or deletion. The approach selected here
is to only store new data on insert, old data on deletion and
differences on update.

## Example

The following assumes you have access to a database named `test` and the
ability to make a new schema called `audit`:

    test=> create schema audit;
    test=> set search_path = audit;
    test=> \i logger.sql 

    CREATE TABLE example (
        id SERIAL PRIMARY KEY
        ,t_stuff TEXT
        ,b_switch BOOLEAN NOT NULL DEFAULT FALSE
    );

    CREATE TRIGGER
        tr_log
    AFTER INSERT OR UPDATE OR DELETE ON 
        example
    FOR EACH ROW EXECUTE PROCEDURE
        fn_logger();

    test=> insert into example (t_stuff, b_switch) values ('hi', false) returning *;
    test=> insert into example (t_stuff, b_switch) values ('there', true) returning *;
    test=> update example set t_stuff = t_stuff || '_updated' where id = 1 returning *;
    test=> delete from example where id = 2 returning *;
    test=> select * from log;

     id | ...  dt_modified ... | e_action | t_table | n_id |                    j_changes                    
    ----+-...--------------...-+----------+---------+------+-------------------------------------------------
      1 | ...-20 20:36:11.3... | insert   | example |    1 | {"id": 1, "t_stuff": "hi", "b_switch": false}
      2 | ...-20 20:36:29.2... | insert   | example |    2 | {"id": 2, "t_stuff": "there", "b_switch": true}
      3 | ...-20 20:36:55.3... | update   | example |    1 | {"t_stuff": "hi_updated"}
      4 | ...-20 20:37:16.6... | delete   | example |    2 | {"id": 2, "t_stuff": "there", "b_switch": true}

An example of registering a trigger to filter out a column from logging.

    CREATE TABLE example2 (
        id SERIAL PRIMARY KEY
        ,t_stuff TEXT
        ,b_switch BOOLEAN NOT NULL DEFAULT FALSE
    );

    CREATE TRIGGER
        tr_log
    AFTER INSERT OR UPDATE OR DELETE ON 
        example2
    FOR EACH ROW EXECUTE PROCEDURE
        -- don't log the 'b_switch' column
        fn_logger('b_switch');

    test=> insert into example2 (t_stuff, b_switch) values ('example2', true) returning *;
    -- this change won't register
    test=> update example2 set b_switch = false where id = 1 returning *;
    -- but this one will
    test=> update example2 set t_stuff = 'change' where id = 1;
    test=> select * from log where t_table = 'example2' and n_id = 1 order by id;

     id | ... dt_modified ... | e_action | t_table  | n_id |            j_changes             
    ----+-...-------------...-+----------+----------+------+----------------------------------
      5 | ...20 20:46:20.0... | insert   | example2 |    1 | {"id": 1, "t_stuff": "example2"}
      6 | ...20 20:52:37.2... | update   | example2 |    1 | {"t_stuff": "change"}

    test=> drop schema audit cascade;

