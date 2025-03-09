#!/bin/bash

AUTH_TOKEN="test-it-settr" # config.yml -> app.auth_token value
HOSTNAME="http://127.0.0.1:8880"

email="admin@localhost"
text="You're invited to check incident"

curl -s -H "Authorization: Bearer $AUTH_TOKEN" \
    --data-urlencode "email=$email" \
    --data-urlencode "message=$text" \
    "${HOSTNAME}/api/v1/make_call" | jq
