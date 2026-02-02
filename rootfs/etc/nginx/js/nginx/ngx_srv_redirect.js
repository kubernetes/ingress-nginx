function srv_redirect(req) {
    const redirectTo = req.variables.tmp_redirect_to;

    const requestUri = req.variables.request_uri.replace(/\/$/, '');

    const useForwardedHeaders = req.variables.forwarded_headers
    const xForwardedProto = req.variables.http_x_forwarded_proto;
    const xForwardedPort = req.variables.http_x_forwarded_port;

    const redirectScheme = useForwardedHeaders && xForwardedProto ? xForwardedProto : req.variables.scheme;
    const redirectPort = useForwardedHeaders && xForwardedPort ? xForwardedPort : req.variables.server_port;

    return `${redirectScheme}://${redirectTo}:${redirectPort}${requestUri}`;
}

export default { srv_redirect };
