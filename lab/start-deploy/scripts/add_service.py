#!/usr/bin/env python3
# pylint: disable=R0903,C0111

import sys
import requests

APIGATEWAY = "http://127.0.0.1:10080"
APP_ID = "c025cb97-9cba-4a3d-aff9-7c5a66c967aa"
APP_SECRET = "XUEJEGzQyeKSKYCKOSbdmaxhRuEzcJxQ"


class Service:

    def __init__(self, app_id, app_secret):
        self.app_id = app_id
        self.app_secret = app_secret
        self._access_token = None
        self.apigateway = APIGATEWAY

    def get_url(self, url):
        return f"{self.apigateway}{url}"

    def get_authn_url(self, url):
        return f"{self.apigateway}/authn{url}"

    def get_service_url(self, url):
        return f"{self.apigateway}/service{url}"

    @property
    def access_token(self):
        if not self._access_token:
            resp = requests.post(self.get_authn_url("/app_token"), json={
                "app_id": self.app_id,
                "app_secret": self.app_secret})
            body = resp.json()
            self._access_token = body["data"]["access_token"]
        return self._access_token

    def add_service(self, name, spec):
        multipart_form_data = {
            'openapi': (spec, open(spec, 'rb')),
            'name': ('', name),
        }
        resp = requests.post(
            self.get_service_url("/service"),
            headers={"Authorization": f"Bearer {self.access_token}"},
            files=multipart_form_data)
        print(resp.text)


def main():
    srv = Service(app_id=APP_ID, app_secret=APP_SECRET)
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} SERVICE_NAME SERVICE_SPEC")
        sys.exit(1)

    name = sys.argv[1]
    spec = sys.argv[2]
    srv.add_service(name, spec)


if __name__ == '__main__':
    main()
