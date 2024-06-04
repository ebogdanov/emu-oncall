#!/bin/bash

HOSTNAME="http://127.0.0.1:8880"
curl -s -H "Authorization: Bearer test-it-settr" -X POST "${HOSTNAME}/api/v1/integrations/" | jq