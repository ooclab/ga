#! /bin/bash
ETCDSET="../ga etcd set"
OPENAPI_SPEC=~/data/projects/github/ooclab/authz/src/codebase/schema.yml
$ETCDSET /ga/service/authz/openapi/spec ${OPENAPI_SPEC} --value-is-file
