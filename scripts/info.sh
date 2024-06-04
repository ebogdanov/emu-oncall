#!/bin/bash

HOSTNAME="http://127.0.0.1:10080"

curl "${HOSTNAME}/api/v1/info" | jq

