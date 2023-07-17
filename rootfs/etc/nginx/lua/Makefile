 all:
	@echo --- build
	@echo CFLAGS: $(CFLAGS)
	@echo LIBFLAG: $(LIBFLAG)
	@echo LUA_LIBDIR: $(LUA_LIBDIR)
	@echo LUA_BINDIR: $(LUA_BINDIR)
	@echo LUA_INCDIR: $(LUA_INCDIR)

.PHONY: install
install:
	@echo $(INST_PREFIX)
	@echo $(INST_CONFDIR)
	@echo INST_LIBDIR: $(INST_LIBDIR)
	@echo INST_LUADIR: $(INST_LUADIR)
	mkdir $(INST_PREFIX)/ingress-nginx-lua
	cp -r . $(INST_PREFIX)/ingress-nginx-lua
	rm -rf $(INST_PREFIX)/ingress-nginx-lua/test $(INST_PREFIX)/ingress-nginx-lua/Makefile $(INST_PREFIX)/ingress-nginx-lua/*.rockspec