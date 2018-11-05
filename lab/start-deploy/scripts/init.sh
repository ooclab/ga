#! /bin/bash
#ETCDSET="../ga etcd set"
#VALUE_FILE=../lab/start-deploy/config/authn/keys/public_key.pem
#$ETCDSET /ga/middleware/uid/public_key ${VALUE_FILE} --value-is-file

APP_ID=c13ec57b-2af8-4884-827e-dee5a6a6f841

export ETCDCTL_API=3
cat ../config/authn/keys/public_key.pem | etcdctl put /ga/middleware/uid/public_key
cat ../services/authn/schema.yml | etcdctl put /ga/service/authn/openapi/spec
cat ../services/authz/schema.yml | etcdctl put /ga/service/authz/openapi/spec
cat ../services/service/schema.yml | etcdctl put /ga/service/service/openapi/spec

# add roles to permission
etcdctl put ga.auth.permissions.authn:post:/app_token.roles '["authenticated"]'
etcdctl put ga.auth.permissions.service:post:/service.roles '["admin"]'

# add roles to user
etcdctl put ga.auth.users.${APP_ID}.roles '["admin"]'
