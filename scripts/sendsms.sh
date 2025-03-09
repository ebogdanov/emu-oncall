#!/bin/bash

AUTH_TOKEN="test-it-settr" # config.yml -> app.auth_token value
HOSTNAME="http://localhost:8880"

text="You are invited to check an incident #13 with title \"ServiceUnreachable\", alert channel: AlertManager, alerts registered: 1, https://grafana-new.biz-systems.ru/a/grafana-oncall-app/alert-groups/I9BG3D69SI48H Your Grafana OnCall <3"
email="admin@localhost"

curl -s -H "Authorization: Bearer $AUTH_TOKEN" -s \
    --data-urlencode "email=${email}" \
    --data-urlencode "message=${text}" \
    "${HOSTNAME}/api/v1/send_sms/"
