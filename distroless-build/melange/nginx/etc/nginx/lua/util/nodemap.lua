local math_random = require("math").random
local util_tablelength = require("util").tablelength
local ngx = ngx
local pairs = pairs
local string = string
local setmetatable = setmetatable

local _M = {}

--- create_map generates the node hash table
-- @tparam {[string]=number} nodes A table with the node as a key and its weight as a value.
-- @tparam string salt A salt that will be used to generate salted hash keys.
local function create_map(nodes, salt)
  local hash_map = {}

  for endpoint, _ in pairs(nodes) do
    -- obfuscate the endpoint with a shared key to prevent brute force
    -- and rainbow table attacks which could reveal internal endpoints
    local key = salt .. endpoint
    local hash_key = ngx.md5(key)
    hash_map[hash_key] = endpoint
  end

  return hash_map
end

--- get_random_node picks a random node from the given map.
-- @tparam {[string], ...} map A key to node hash table.
-- @treturn string,string The node and its key
local function get_random_node(map)
  local size = util_tablelength(map)

  if size < 1 then
    return nil, nil
  end

  local index = math_random(1, size)
  local count = 1

  for key, endpoint in pairs(map) do
      if count == index then
        return endpoint, key
      end

      count = count + 1
  end

  ngx.log(ngx.ERR, string.format("Failed to find node %d of %d! "
    .. "This is a bug, please report!", index, size))

  return nil, nil
end

--- new constructs a new instance of the node map
--
-- The map uses MD5 to create hash keys for a given node. For security reasons it supports
-- salted hash keys, to prevent attackers from using rainbow tables or brute forcing
-- the node endpoints, which would reveal cluster internal network information.
--
-- To make sure hash keys are reproducible on different ingress controller instances the salt
-- needs to be shared and therefore is not simply generated randomly.
--
-- @tparam {[string]=number} endpoints A table with the node endpoint
-- as a key and its weight as a value.
-- @tparam[opt] string hash_salt A optional hash salt that will be used to obfuscate the hash key.
function _M.new(self, endpoints, hash_salt)
  if hash_salt == nil then
    hash_salt = ''
  end

  -- the endpoints have to be saved as 'nodes' to keep compatibility to balancer.resty
  local o = {
    salt = hash_salt,
    nodes = endpoints,
    map = create_map(endpoints, hash_salt)
  }

  setmetatable(o, self)
  self.__index = self
  return o
end

--- reinit reinitializes the node map reusing the original salt
-- @tparam {[string]=number} nodes A table with the node as a key and its weight as a value.
function _M.reinit(self, nodes)
  self.nodes = nodes
  self.map = create_map(nodes, self.salt)
end

--- find looks up a node by hash key.
-- @tparam string key The hash key.
-- @treturn string The node.
function _M.find(self, key)
  return self.map[key]
end

--- random picks a random node from the hashmap.
-- @treturn string,string A random node and its key or both nil.
function _M.random(self)
  return get_random_node(self.map)
end

--- random_except picks a random node from the hashmap, ignoring the nodes in the given table
-- @tparam {string, } ignore_nodes A table of nodes to ignore, the node needs to be the key,
--                                 the value needs to be set to true
-- @treturn string,string A random node and its key or both nil.
function _M.random_except(self, ignore_nodes)
  local valid_nodes = {}

  -- avoid generating the map if no ignores where provided
  if ignore_nodes == nil or util_tablelength(ignore_nodes) == 0 then
    return get_random_node(self.map)
  end

  -- generate valid endpoints
  for key, endpoint in pairs(self.map) do
      if not ignore_nodes[endpoint] then
        valid_nodes[key] = endpoint
      end
  end

  return get_random_node(valid_nodes)
end

return _M
