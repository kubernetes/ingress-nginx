function handle_cors(req) {
    const originsRegex = new RegExp(`${req.variables.cors_origins_regex}$`, 'i');

    if (originsRegex.test(req.headersIn['Origin'])) {
        const allowedOrigins = req.variables.cors_allowed_origins.split(',');

        req.headersOut['Access-Control-Allow-Origin'] = allowedOrigins.length === 1 && allowedOrigins[0] === '*' ? '*' : req.headersIn['Origin'];
        req.headersOut['Access-Control-Allow-Methods'] = req.variables.cors_allow_methods;
        req.headersOut['Access-Control-Allow-Headers'] = req.variables.cors_allow_headers;
        req.headersOut['Access-Control-Max-Age'] = req.variables.cors_max_age;
        if (req.variables.cors_allow_credentials) req.headersOut['Access-Control-Allow-Credentials'] = req.variables.cors_allow_credentials;
        if (req.variables.cors_expose_headers) req.headersOut['Access-Control-Expose-Headers'] = req.variables.cors_expose_headers;

        if (req.method === 'OPTIONS') {
            req.headersOut['Content-Type'] = 'text/plain charset=UTF-8';
            req.headersOut['Content-Length'] = '0';
        }
    }
}

export default {handle_cors};
