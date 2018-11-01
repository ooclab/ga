package serve

var yamlConfigExample = []byte(`version: "1"

# use the service and config
services:

    etcd:
        endpoints:
            - etcd:2379

    authn:
        baseurl: http://traefik/authn
        app_id: xxx
        app_secret: xxx


# start the forward server
servers:

    # forward the request from backend service to others (e.g. traefik)
    internal:
        type: http
        listen: ":2998"
        backend: http://traefik:10080
        middlewares:
            authadd: {}

    # forward the request from api gateway to backend service (e.g. traefik)
    external:
        type: http
        listen: ":2999"
        backend: http://api:3000
        middlewares:
            uid:
                public_key_etcd: ""
            authz:
                service_name: "myservice"
                openapi_sepc_etcd: "/ga/service/myservice/openapi/spec"
`)
