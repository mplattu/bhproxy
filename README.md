# bhproxy - Behold.so proxy

This is a proxy between HTML5 application and Behold.so web service.

It simplifies JSON object served by Behold and takes load from the service.

The bhproxy lives behind a web service via CGI which makes possible to run it in a regular LAMP web server.

## Configuration

bhproxy reads following environment variables:

* `DB_FILENAME` - a path to rw file used to store SQLite database.

## Developing

* Build: `make build` or `make build-dev` creates a binary `cgi-bin/bhproxy`
* Try: `make start` creates a Python3 web server. The binary answers at http://localhost:8080/cgi-bin/bhproxy?id=BEHOLD_FEED_ID
* To run tests: `make test` or `make test-v`
