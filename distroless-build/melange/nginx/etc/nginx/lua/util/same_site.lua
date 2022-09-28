local string = string

local _M = {}

-- determines whether to apply a SameSite=None attribute
-- to a cookie, based on the user agent.
-- returns: boolean
--
-- Chrome 80 treating third-party cookies as SameSite=Strict
-- if SameSite is missing. Certain old browsers don't recognize
-- SameSite=None and will reject cookies entirely bearing SameSite=None.
-- This creates a situation where fixing things for
-- Chrome >= 80 breaks things for old browsers.
-- This function compares the user agent against known
-- browsers which will reject SameSite=None cookies.
-- reference: https://www.chromium.org/updates/same-site/incompatible-clients
function _M.same_site_none_compatible(user_agent)
  if not user_agent then
    return true
  elseif string.match(user_agent, "Chrome/4") then
    return false
  elseif string.match(user_agent, "Chrome/5") then
    return false
  elseif string.match(user_agent, "Chrome/6") then
    return false
  elseif string.match(user_agent, "CPU iPhone OS 12") then
    return false
  elseif string.match(user_agent, "iPad; CPU OS 12") then
    return false
  elseif string.match(user_agent, "Macintosh")
      and string.match(user_agent, "Intel Mac OS X 10_14")
      and string.match(user_agent, "Safari")
      and not string.match(user_agent, "Chrome") then
    return false
  end

  return true
end

return _M
