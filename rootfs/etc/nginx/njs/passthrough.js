function truncate(r) {
    try {
        ngx.shared.ptbackends.clear() // TODO: We should instead try to compare and clean
        r.return(200, "ok")
        r.finish()
    } catch (e) {
        r.error(e)
        r.return(400, "error truncating the map json payload")
        r.finish()
    }
}

function set(r) {
    var service;
    service = r.args.key
    if (service == "" || service == null ) {
        r.return(400, "key should not be null")
        r.finish()
        return
    }
    
    try {
        JSON.parse(r.requestText)
        ngx.shared.ptbackends.set(r.args.key, r.requestText)
        r.return(200, "ok")
        r.finish()
    } catch (e) {
        r.error(e)
        r.return(400, "error parsing json payload")
        r.finish()
    }
}

function getUpstream(r) {
    var service;

    try {
        if ("variables" in r) {
            service = r.variables.ssl_preread_server_name;
        }

        if (service == "") {
            // TODO: This should be a parameter with the port that NGINX is listening
            // for non Passthrough
            return "127.0.0.1:442"
        }
    
        const backends = ngx.shared.ptbackends.get(service)
        if (backends == "" || backends == null) {
            throw "no backend configured"
        }

        const objBackend = JSON.parse(backends)
        if (objBackend["endpoints"] == null || objBackend["endpoints"] == undefined) {
            throw "bad endpoints object" // TODO: This validation should happen when receiving the json
        }

        // TODO: We can loadbalance between backends, but right now let's receive just the ClusterIP
        if (!Array.isArray(objBackend["endpoints"])) {
            throw "endpoint object is not an array"
        }

        if (objBackend["endpoints"].length < 1) {
            throw "no backends available for the service"
        }

        // TODO: Do we want to implement a different LB for Passthrough when it is composed of multiple backends?
        var randomBackend = Math.floor(Math.random() * (objBackend["endpoints"].length));
        if (typeof objBackend["endpoints"][randomBackend] != 'string') {
            throw "endpoint is not a string"        
        }
        return objBackend["endpoints"][randomBackend]
    
    } catch (e) {
        // If there's an error we should give user a return saying it
        return "@invalidbackend"
    } 
}

export default {set, truncate, getUpstream};