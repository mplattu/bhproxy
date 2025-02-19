TEMP_DIR := $(TMPDIR)
ifeq ($(TEMP_DIR),)
    TEMP_DIR := $(TEMP)
endif
ifeq ($(TEMP_DIR),)
    TEMP_DIR := /tmp
endif

DB_FILENAME=$(TEMP_DIR)/bhproxy.sqlite

start:
	if [ ! -f $(DB_FILENAME) ]; then touch $(DB_FILENAME); fi
	DB_FILENAME=$(DB_FILENAME) python3 -m http.server --bind localhost --cgi 8080

start-minihttpd:
	if [ -f /tmp/mini_httpd.log ]; then rm /tmp/mini_httpd.log; fi
	if [ -f /tmp/mini_httpd.pid ]; then rm /tmp/mini_httpd.pid; fi
	mini_httpd -p 8080 -c "cgi-bin/**"  -l /tmp/mini_httpd.log -i /tmp/mini_httpd.pid

stop-minihttpd:
	kill -TERM `cat /tmp/mini_httpd.pid`

build:
	@CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o cgi-bin/bhproxy cmd/main.go

build-dev:
	@go build -o cgi-bin/bhproxy cmd/main.go

test:
	@go test ./pkg/feed
	@go test ./pkg/utility
	@go test ./pkg/db
	@go test ./pkg/handler

test-v:
	@go test -v ./pkg/feed
	@go test -v ./pkg/utility
	@go test -v ./pkg/db
	@go test -v ./pkg/handler
