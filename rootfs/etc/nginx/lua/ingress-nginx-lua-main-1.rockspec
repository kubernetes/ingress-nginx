package = "ingress-nginx-lua"
version = "main-1"
supported_platforms = {"linux", "macosx"}
source = {
   url = "git+https://github.com/kubernetes/ingress-nginx.git"
}
description = {
   summary = "Lua modules for ingress-nginx.",
   homepage = "https://github.com/kubernetes/ingress-nginx",
   license = "Apache License 2.0"
}
dependencies = {
   "lua-resty-global-throttle >= 0.2"
}
build = {
   type = "make",
   build_variables = {
      CFLAGS="$(CFLAGS)",
      LIBFLAG="$(LIBFLAG)",
      LUA_LIBDIR="$(LUA_LIBDIR)",
      LUA_BINDIR="$(LUA_BINDIR)",
      LUA_INCDIR="$(LUA_INCDIR)",
      LUA="$(LUA)",
   },
   install_variables = {
      INST_PREFIX="$(PREFIX)",
      INST_BINDIR="$(BINDIR)",
      INST_LIBDIR="$(LIBDIR)",
      INST_LUADIR="$(LUADIR)",
      INST_CONFDIR="$(CONFDIR)",
   },
}
