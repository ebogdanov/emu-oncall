#!/bin/bash

HOSTNAME="http://127.0.0.1:10080"

curl -s "${HOSTNAME}/api/v1/users?page=1&short=false&roles=0&roles=1" | jq
