/* This is the passthrough configuration endpoint
It will work as following:
* Ingress controller calls the configEndpoint with a full json of hosts, 
    endpoints
* This json will be parsed and each key will be added as the host, and the 
    config will be a json containing the endpoint
* Once a request arrive on the public endpoint (port 443), the getPTBackend 
    function will be called (as part of a js_set, and return the right backend)
  * It will read the host from variables.ssl_preread_server_name
  * It will get the config from the shared map
  * PROXY protocol is not supported today, as NGINX does not allows us to set 
    conditional proxy with a variable
  * It will return the backend as a single value
* We don't support multiple backends right now

expectedJson = `
{
    "server1.tld": {
        "endpoint": "10.20.30.40:12345",
    },
    "server2.tld": {
        "endpoint": "10.20.30.50:8080",
    },
}
`
*/

/**
configstatus will be the variable that will contain the latest configuration status return, 
so the caller will be able to check it. It is returned by getConfigStatus, which is the last
operation on the configuration endpoint
*/

const OK = "OK";
const NOK = "NOK";   
const KEYNAME = "passthroughmap";

var configstatus = ''
function getConfigStatus() {
    return configstatus;
}

/**
configPTBackends will receive a JSON as defined above, and:
  * Parse the input to validate if it is valid
  * Truncate the previous map
  * Configure with this new structure
We need to be careful here, if some situation where a config is ongoing may break a starting communication. Something we could look at
is to put a lock somewhere, maybe on a different shared map, and release the connection just once it is unlocked
*/
function configBackends(s) {
    s.log("Start of configuration");
    var req = '';
    s.on('upload', function(data, flags) {
        // so far, we just want to receive and store the data until the stream finishes
        req += data;  
        if (data.length || flags.last) {
            configstatus = configureWithData(req, s)
            s.warn(configstatus)
            s.done();
        }
    }) 
}

function configureWithData(configdata, s) {
    try {
        let backends = {};
        const parsed = JSON.parse(configdata);
        const keys = Object.keys(parsed);
        
        // We create a separate array/object as a safety measure, so if something is broken
        // it will not break the whole reconfiguration
        keys.forEach((key) => {
            let serviceitem = parsed[key];
            if (typeof serviceitem.endpoint != "string") {
                s.warn(`endpoint of ${key} is not string, skipping`)
                return;
            }
            backends[key] = serviceitem;
        });

        // Clear method is not working, we should verify with NGX folks 
        //ngx.shared.passthrough.clear();
        ngx.shared.ptbackends.set(KEYNAME, JSON.stringify(backends))

        return OK
    } catch (e) {
        s.error(`failed configuring data: ${e}`);
        return NOK;
    }
}

const PROXYSOCKET="unix:/var/run/nginx/streamproxy.sock";
// getBackend fetches the backend given a hostname sent via SNI
function getBackend(s) {
    try {
        const backendCfg = getBackendEndpoint(s);
        if(backendCfg[1]) {
            return PROXYSOCKET
        }
        return backendCfg[0]
    } catch(e) {
        s.warn(`error occurred while getting the backend ` +
        `sending to default backend: ${e}`)
    
        return "127.0.0.1:442"
    }
}

// getProxiedBackend fetches the backend given a hostname sent via SNI, to be used by proxy_protocol endpoint.
// An error here should be a final error
function getProxiedBackend(s) {
    try {
        const backend = getBackendEndpoint(s)[0];
        return backend;

    } catch(e) {
        s.warn(`error occurred while getting the backend ` +
        `sending to default backend: ${e}`) 
        s.deny()
    }
}

// getBackendEndpoint is the common function to return the endpoint and optinally if it should
// use proxy_protocol from the map
function getBackendEndpoint(s) {
    var hostname = s.variables.ssl_preread_server_name;
    if (hostname == null || hostname == "undefined" || hostname == "") {
        throw("hostname was not provided")
    }

    let backends = ngx.shared.ptbackends.get(KEYNAME)
    if (backends == null || backends == "") {
        throw('no entry on endpoint map')
    }
    const backendmap = JSON.parse(backends)
    if (backendmap[hostname] == null || backendmap[hostname] == undefined || 
            backendmap[hostname].endpoint == null || backendmap[hostname].endpoint == undefined) {
        throw `no endpoint is configured for service ${hostname}"`
    }

    var isProxy = false
    if (typeof backendmap[hostname].use_proxy == "boolean" && backendmap[hostname].use_proxy) {
        isProxy = backendmap[hostname].use_proxy
    }

    return [backendmap[hostname].endpoint, isProxy];
}

export default {getConfigStatus, configBackends, getBackend, getProxiedBackend};
