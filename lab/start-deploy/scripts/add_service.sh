#! /bin/bash
# ETCDSET="../ga etcd set"
# #OPENAPI_SPEC=~/data/projects/github/ooclab/authz/src/codebase/schema.yml
# SERVICE=authz
# OPENAPI_SPEC=../lab/start-deploy/config/authz/schema.yml
# $ETCDSET /ga/service/${SERVICE}/openapi/spec ${OPENAPI_SPEC} --value-is-file

ACCESS_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IjgzMDc0MDJCLTRERUMtNDFFRi1CNzdCLUM5MEZDNUZENzdFNyJ9.eyJzdWIiOiJvb2NsYWIiLCJleHAiOjE1NDM5ODcwNTgsImlhdCI6MTU0MTM5NTA1OCwiaXNzIjoib29jbGFiIiwidWlkIjoiYzEzZWM1N2ItMmFmOC00ODg0LTgyN2UtZGVlNWE2YTZmODQxIn0.a9Sat56xaCKum1hFz8r_fnEPjzTr_ytdZOc-x6frXZ85z1YP0r7ps1NuLIPjX3J8e_VhLkqIMe2lqOlLs0K5V9SctAlF6YBFDIaUz_xtsT3lQQqA6FL3KyAomnunpTL5z1kId-fKLTkWLQc7LbREBU0fcYTCUS5Yngoy4tpuN1vQJe4tM7CqDbe4vGOWFmPKzvTwHJa_uxxn2DUHaNoanQiHE4dyN7RP400Asuxvklxm6I_vtkUN8AYC9LJ6FxhoGxzNO8A_AmL9vU6iH56Jz5INtCjo5uHYGmq4kuC5cUNff2y-BHoHV4HSvgACXQXqv-VP9M62KAL8zNsqCTtrUg
APIGATEWAY=http://127.0.0.1:10080

curl -X POST ${APIGATEWAY}/service/service -F "openapi=@../config/authz/schema.yml" -F "name=authz" -H "Authorization: Bearer ${ACCESS_TOKEN}"
