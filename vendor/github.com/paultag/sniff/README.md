sniff
=====

SNIff is a small server that will accept incoming TLS connections, and parse
TLS Client Hello messages for the SNI Extension. If one is found, we'll go
ahead and forward that connection to a remote (or local!) host.

sniff config
------------

```json
{
    "bind": {
        "host": "localhost",
        "port": 8443
    },
    "servers": [
        {
            "default": false,
            "regexp": false,
            "host": "97.107.130.79",
            "names": [
                "pault.ag",
                "www.pault.ag"
            ],
            "port": 443
        }
    ]
}
```

The following config will listen on port `8443`, and connect any requests
to `pault.ag` or `www.pault.ag` to port `443` on host `97.107.130.79`. If
nothing matches this, the socket will be closed.

Changing default to true would send any unmatched hosts (or TLS / SSL connections
without SNI) to that host.

By default, the requested domain name is compared literally with the strings
inside `names`. If `regexp` is true, then the names are interpreted as regular
expressions. Each server and name will be checked in the order they appear in
the file, stopping with the first match. If there is no match, then the
request is sent to the first server with `default` set.

using the parser
----------------

```
import (
    "fmt"

    "pault.ag/go/sniff/parser"
)

func main() {
    listener, err := net.Listen("tcp", "localhost:2222")
    if err != nil {
        return err
    }
}
```
