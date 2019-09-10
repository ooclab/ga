package serve

var yamlConfigExample = []byte(`version: "2"

# start the forward server
listen: ":2999"

services:

  S1:
    path_prefix: /httpbin
    backend: http://httpbin.ooclab.com
    middlewares:
    - name: logger
    - name: debug
  # forward the request from api gateway to backend service
  S2:
    path_prefix: /api/bfmanage/v1
    backend: http://localhost:3002/bfmanage/v1
    middlewares:
    - name: logger
    - name: debug
    - name: openapi3
      service_name: api
      service_spec: "http://localhost:3002/bfmanage/v1"
`)
