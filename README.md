# bhproxy - Behold.so proxy

This is a proxy between HTML5 application and Behold.so web service.

It simplifies JSON object served by Behold and takes load from the service.

The bhproxy lives behind a web service via CGI which makes possible to run it in a regular LAMP web server.

## Configuration

bhproxy reads following environment variables:

* `BHP_DB_FILENAME` - a path to rw file used to store SQLite database. Required.
* `BHP_IMAGE_DIRECTORY` - a rw path to store all images a without trailing slash. Required.
* `BHP_IMAGE_URL` - prefix for image files located in `IMAGE_DIRECTORY` without a trailing slash. Optional, defaults to root (`/`).
* `BHP_ALLOWED_FEED_IDS` - comma-separated list of Behold feed IDs which this proxy serves. Optional, defaults to all IDs are allowed.
* `BHP_LOGFILE` - path to log file. Optional, defaults to STDERR.

The environment variables can be set using a standard `.env` file which should be in the same directory with the executable.

## Developing

* Build: `make build` or `make build-dev` creates a binary `bin/bhproxy`
* Try: `make start` creates a Python3 web server. The binary answers at http://localhost:8080/cgi-bin/bhproxy?id=BEHOLD_FEED_ID
* To pass `BHP_ALLOWED_FEED_IDS` whitelist: `BHP_ALLOWED_FEED_IDS=JYK0zcST7PconDbzq1GL,JYK0bzSTZPConDbzq1XP make start`
* To run tests: `make test` or `make test-v`
