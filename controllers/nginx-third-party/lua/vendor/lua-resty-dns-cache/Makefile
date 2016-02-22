OPENRESTY_PREFIX=/usr/local/openresty

PREFIX ?=          /usr/local
LUA_INCLUDE_DIR ?= $(PREFIX)/include
LUA_LIB_DIR ?=     $(PREFIX)/lib/lua/$(LUA_VERSION)
INSTALL ?= install
TEST_FILE ?= t

.PHONY: all test leak

all: ;


install: all
	$(INSTALL) -d $(DESTDIR)/$(LUA_LIB_DIR)/resty/dns

leak: all
	TEST_NGINX_CHECK_LEAK=1	TEST_NGINX_NO_SHUFFLE=1 PATH=$(OPENRESTY_PREFIX)/nginx/sbin:$$PATH prove -I../test-nginx/lib -r $(TEST_FILE)

test: all
	TEST_NGINX_NO_SHUFFLE=1 PATH=$(OPENRESTY_PREFIX)/nginx/sbin:$$PATH prove -I../test-nginx/lib -r $(TEST_FILE)
	util/lua-releng.pl

