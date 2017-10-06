#!/bin/sh

set -u


DIR="$(dirname $0)"
CMD="${DIR}/../cert-chain-resolver"
TEMP_FILE="$(mktemp)"


(
    set -e

    # it should output certificate bundle in PEM format with certificate from Comodo, PEM leaf, 2x DER intermediate
    $CMD -o "$TEMP_FILE" "$DIR/comodo.crt"
    diff "$TEMP_FILE" "$DIR/comodo.bundle.crt"

    # it should output certificate bundle in PEM format with certificate from Comodo, DER leaf, 2x DER intermediate
    $CMD -o "$TEMP_FILE" "$DIR/comodo.der.crt"
    diff "$TEMP_FILE" "$DIR/comodo.bundle.crt"

    # it should output certificate bundle in PEM format with certificate from GoDaddy, PEM leaf, PEM intermediate
    $CMD -o "$TEMP_FILE" "$DIR/godaddy.crt"
    diff "$TEMP_FILE" "$DIR/godaddy.bundle.crt"

    # it should output certificate bundle in PEM format with certificate with multiple issuer URLs
    $CMD -o "$TEMP_FILE" "$DIR/multiple-issuer-urls.crt"
    diff "$TEMP_FILE" "$DIR/multiple-issuer-urls.bundle.crt"

    # it should output certificate bundle in DER format
    $CMD -d -o "$TEMP_FILE" "$DIR/comodo.crt"
    diff "$TEMP_FILE" "$DIR/comodo.bundle.der.crt"

    # it should output certificate chain in PEM format
    $CMD -i -o "$TEMP_FILE" "$DIR/comodo.crt"
    diff "$TEMP_FILE" "$DIR/comodo.chain.crt"

    # it should output certificate chain in DER format
    $CMD -d -i -o "$TEMP_FILE" "$DIR/comodo.crt"
    diff "$TEMP_FILE" "$DIR/comodo.chain.der.crt"

    # it should output certificate bundle in PEM format, with input from stdin and output to stdout
    $CMD < "$DIR/comodo.crt" > "$TEMP_FILE"
    diff "$TEMP_FILE" "$DIR/comodo.bundle.crt"

    # Append CA root cert to output
    $CMD -s < "$DIR/comodo.crt" > "$TEMP_FILE"
    diff "$TEMP_FILE" "$DIR/comodo.withca.crt"

    # Already has CA root cert
    $CMD -s < "$DIR/multiple-issuer-urls.crt" > "$TEMP_FILE"
    diff "$TEMP_FILE" "$DIR/multiple-issuer-urls.withca.bundle.crt"

    # DST Root CA X3, PKCS#7 package
    $CMD < "$DIR/dstrootcax3.p7c" > "$TEMP_FILE"
    diff "$TEMP_FILE" "$DIR/dstrootcax3.pem"

    # it should detect invalid certificate
    (
         set +e
         ! echo "xxx" | $CMD
    )

    # Build and start the webserver to serve the certificates
    go build -o  "${DIR}/http-server" "${DIR}/http-server.go"
    sudo "${DIR}/http-server" "${DIR}" &
    PID="$!"
    sleep 1

    # It should correctly detect root certificates to prevent infinite traversal loops when the root
    # certificate also has an AIA Certification Authority Issuer record
    $CMD < "$DIR/self-issued.crt" > "$TEMP_FILE"
    diff "$TEMP_FILE" "$DIR/self-issued.bundle.crt"

    # Stop the webserver
    sudo kill "$PID"
)
STATUS="$?"


rm -f "$TEMP_FILE"

exit "$STATUS"
