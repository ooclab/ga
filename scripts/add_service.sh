#! /bin/bash
ETCDSET="../ga etcd set"
#OPENAPI_SPEC=~/data/projects/github/ooclab/authz/src/codebase/schema.yml
SERVICE=authz
OPENAPI_SPEC=../lab/start-deploy/config/authz/schema.yml
$ETCDSET /ga/service/${SERVICE}/openapi/spec ${OPENAPI_SPEC} --value-is-file
