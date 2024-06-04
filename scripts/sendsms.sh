#!/bin/bash

text="You are invited to check an incident #13 with title \"ServiceUnreachable\", alert channel: AlertManager, alerts registered: 1, https://grafana-new.biz-systems.ru/a/grafana-oncall-app/alert-groups/I9BG3D69SI48H Your Grafana OnCall <3"
email="admin@localhost"

HOSTNAME="http://localhost:8880"

curl -s -H "Authorization: Bearer test-it-settr" -s \
    --data-urlencode "email=${email}" \
    --data-urlencode "message=${text}" \
    "${HOSTNAME}/api/v1/send_sms/" | jq