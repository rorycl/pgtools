# pgtools/modelmaker

A Go tool for creating model files for Postgresql plpgsql functions.

Rory Campbell-Lange  
version v0.1-beta : 21 March 2021

## Introduction

Inspired by the work of a colleague who has written a python programme
for automatically creating model interface files, this tool helps
introspect the functions in a particular database and output these for
easy inclusion in a model file, for example by running the programme via
`:r ! modelmaker <args>` in vim.

The included output templates are examples for creating snippets for
including in python database model files.

## Example

Run the programme with `-h`:

	Usage:
	  modelmaker : generate python model entries for plpgsql functions

	Application Options:
	  -u, --user=       database user
	  -p, --password=   database pass
	  -d, --database=   database
	  -P, --port=       server port (default: 5432)
	  -H, --host=       server host (default: 127.0.0.1)
	  -s, --searchpath= searchpath
	  -t, --template=   template file (default: output.tpl)
	  -f, --filter=     filter for function names (regexes allowed)

	Help Options:
	  -h, --help        Show this help message

Example output:

	./modelmaker -d testdb -u test -p pass  \
                 -s "testschema, public" \
                 -t python2.tpl \
                 -f creject

Depending on which output template is chosen, you might get output along
the following lines:

    def testschema_creject(self, operator, in_id):
        """
        Namespace: testschema 
        Input parameters:
            operator  : integer default: None
            id        : integer default: None
        Returns:
            type: contracts
        """
        return self.callproc('testschema.creject', (operator, id))

Or, for a different call using the drop function template:

	DROP FUNCTION ctest.fncancel (integer, integer, date, date, text, text);
	DROP FUNCTION ctest.fnreject (integer, integer);
	DROP FUNCTION ctest.fnsign (integer, integer, text, boolean);
	DROP FUNCTION ctest.fn_jsonb_diff (jsonb, jsonb);

The template is a standard go text template.
