#! /bin/bash
ETCDSET="../ga etcd set"
VALUE_FILE=../lab/start-deploy/config/authn/keys/public_key.pem
$ETCDSET /ga/middleware/uid/public_key ${VALUE_FILE} --value-is-file
