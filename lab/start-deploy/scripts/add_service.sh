#! /bin/bash
# ETCDSET="../ga etcd set"
# #OPENAPI_SPEC=~/data/projects/github/ooclab/authz/src/codebase/schema.yml
# SERVICE=authz
# OPENAPI_SPEC=../lab/start-deploy/config/authz/schema.yml
# $ETCDSET /ga/service/${SERVICE}/openapi/spec ${OPENAPI_SPEC} --value-is-file

SERVICE_SERVICE=http://127.0.0.1:3000
curl -X POST ${SERVICE_SERVICE}/service -F "openapi=@../config/authz/schema.yml" -F "name=authz"
