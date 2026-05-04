# Creating the server and client keys and certificates for the test docker HiveMQ (OpenSSL)

- Ref. [HiveMQ: HowTo configure server-side TLS with HiveMQ and Keytool (self-signed)](https://www.hivemq.com/docs/hivemq/4.4/user-guide/howtos.html)
- Ref. [stackoverflow: How to add subject alernative name to ssl certs?](https://stackoverflow.com/questions/8744607/how-to-add-subject-alernative-name-to-ssl-certs#8744717)

------------------------------------------------------------------------

## Create server key and certificate

### 1. Create an OpenSSL configuration file (for SAN support)

Create `server.cnf`:

    [req]
    default_bits       = 2048
    distinguished_name = dn
    req_extensions     = req_ext
    x509_extensions    = req_ext
    prompt             = no

    [dn]
    C  = MQ
    ST = docker
    L  = docker
    O  = uhppoted
    OU = hivemq
    CN = localhost

    [req_ext]
    subjectAltName = @alt_names

    [alt_names]
    DNS.1 = localhost
    IP.1  = 192.168.1.100

------------------------------------------------------------------------

### 2. Generate server key and self-signed certificate

    openssl req -x509 -nodes -days 365 \
      -newkey rsa:2048 \
      -keyout localhost.key \
      -out localhost.pem \
      -config server.cnf

------------------------------------------------------------------------

### 3. Check the server certificate

    openssl x509 -in localhost.pem -noout -text

------------------------------------------------------------------------

### 4. (Optional) Create a PKCS#12 keystore

    openssl pkcs12 -export \
      -in localhost.pem \
      -inkey localhost.key \
      -out localhost.p12 \
      -name hivemq \
      -passout pass:hivemq

------------------------------------------------------------------------

### 5. (Optional) Convert PKCS#12 to JKS

    keytool -importkeystore \
      -srckeystore localhost.p12 \
      -srcstoretype PKCS12 \
      -srcstorepass hivemq \
      -destkeystore localhost.jks \
      -deststorepass hivemq

------------------------------------------------------------------------

### 6. Copy the server key and certificate to the `docker` folders

    cp localhost.key ./docker/hivemq/localhost.key
    cp localhost.pem ./docker/hivemq/localhost.pem
    cp localhost.pem ./docker/uhppoted-mqtt/hivemq.pem

    cp localhost.key ./docker/integration-tests/hivemq/localhost.key
    cp localhost.pem ./docker/integration-tests/hivemq/localhost.pem
    cp localhost.pem ./docker/integration-tests/mqttd/localhost.pem

------------------------------------------------------------------------

### 7. Copy the server certificate to the `uhppoted` configuration folder

    cp localhost.pem /usr/local/etc/com.github.uhppoted/hivemq.pem

------------------------------------------------------------------------

## Create client keys and certificates

### 1. Generate client key pair and self-signed certificate

    openssl req -x509 -nodes -days 365 \
      -newkey rsa:2048 \
      -keyout client.key \
      -out client.cert \
      -subj "/C=MQ/ST=localhost/L=localhost/O=localhost/OU=localhost/CN=localhost"

------------------------------------------------------------------------

### 2. Check the client certificate

    openssl x509 -in client.cert -noout -text

------------------------------------------------------------------------

### 3. Export client certificate as a DER file

    openssl x509 -outform der -in client.cert -out client.crt

------------------------------------------------------------------------

### 4. Create a truststore

    cp client.cert clients.pem

Optional Java truststore:

    keytool -import \
      -file client.crt \
      -alias client \
      -keystore clients.jks \
      -storepass hivemq \
      -noprompt

------------------------------------------------------------------------

### 5. Copy the client key and certificate

    cp client.key  /usr/local/etc/com.github.uhppoted/mqtt/client.key
    cp client.cert /usr/local/etc/com.github.uhppoted/mqtt/client.cert

------------------------------------------------------------------------

### 6. Copy to docker folders

    cp client.key  ~/Development/uhppote/uhppoted/docker/hivemq/client.key
    cp client.key  ~/Development/uhppote/uhppoted/docker/uhppoted-mqtt/secure/client.key
    cp client.key  ~/Development/uhppote/uhppoted/docker/integration-tests/hivemq/client.key
    cp client.key  ~/Development/uhppote/uhppoted/docker/integration-tests/mqttd/secure/client.key

    cp client.cert ~/Development/uhppote/uhppoted/docker/hivemq/client.cert
    cp client.cert ~/Development/uhppote/uhppoted/docker/uhppoted-mqtt/secure/client.cert
    cp client.cert ~/Development/uhppote/uhppoted/docker/integration-tests/hivemq/client.cert
    cp client.cert ~/Development/uhppote/uhppoted/docker/integration-tests/mqttd/secure/client.cert

    cp client.crt  ~/Development/uhppote/uhppoted/docker/hivemq/client.crt
    cp client.crt  ~/Development/uhppote/uhppoted/docker/integration-tests/hivemq/client.crt

------------------------------------------------------------------------

## Rebuild Docker images

    cd ./docker/hivemq        && docker build -f Dockerfile -t hivemq/uhppoted .
    cd ./docker/uhppoted-mqtt && docker build -f Dockerfile -t uhppoted/mqtt .
