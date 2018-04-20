local cwd = io.popen("pwd"):read('*l')
package.path = cwd .. "/rootfs/etc/nginx/lua/?.lua;" .. package.path

describe("[chash_test]", function()
  -- TODO(elvinefendi) add unit tests
end)
