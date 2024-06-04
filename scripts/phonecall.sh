#!/bin/bash

HOSTNAME="http://127.0.0.1:8880"

email="admin@localhost1"

curl --data-urlencode "email=$email" \
    --data-urlencode "message=You're invited to check incident" \
    -H "Authorization: 08c215f69f728eb216346866410239b515be728ef646f2df91052e66673bad98" \
    -v \
    "${HOSTNAME}/api/v1/make_call" | jq
