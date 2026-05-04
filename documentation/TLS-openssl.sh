#/bin/bash

set -e

rm -rf tmp docker etc

mkdir -p ./tmp

# ------------------------------------------------------------------
# Create OpenSSL config (for SAN support)
# ------------------------------------------------------------------
cat > ./tmp/server.cnf <<EOF
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
EOF

# generate server key + certificate
openssl req -x509 -nodes -days 365 \
  -newkey rsa:2048 \
  -keyout ./tmp/localhost.key \
  -out ./tmp/localhost.pem \
  -config ./tmp/server.cnf

# inspect certificate
openssl x509 -in ./tmp/localhost.pem -noout -text

# (optional) Create PKCS#12 keystore
openssl pkcs12 -export \
  -in ./tmp/localhost.pem \
  -inkey ./tmp/localhost.key \
  -out ./tmp/localhost.p12 \
  -name hivemq \
  -passout pass:hivemq

# (optional) Convert to JKS (only if required)
# keytool -importkeystore \
#   -srckeystore ./tmp/localhost.p12 \
#   -srcstoretype PKCS12 \
#   -srcstorepass hivemq \
#   -destkeystore ./tmp/localhost.jks \
#   -deststorepass hivemq

# create directory structure
mkdir -p ./docker/hivemq
mkdir -p ./docker/uhppoted-mqtt
mkdir -p ./docker/integration-tests/hivemq
mkdir -p ./docker/integration-tests/mqttd

# copy server certs/keys
cp ./tmp/localhost.pem ./docker/hivemq/localhost.pem
cp ./tmp/localhost.key ./docker/hivemq/localhost.key

cp ./tmp/localhost.pem ./docker/integration-tests/hivemq/localhost.pem
cp ./tmp/localhost.key ./docker/integration-tests/hivemq/localhost.key

cp ./tmp/localhost.pem ./docker/integration-tests/mqttd/localhost.pem
cp ./tmp/localhost.pem ./docker/uhppoted-mqtt/hivemq.pem

# (optional) keystore copies
cp ./tmp/localhost.p12 ./docker/hivemq/localhost.p12
cp ./tmp/localhost.p12 ./docker/integration-tests/hivemq/localhost.p12

# copy to etc config
mkdir -p ./etc/com.github.uhppoted
cp ./tmp/localhost.pem ./etc/com.github.uhppoted/hivemq.pem
