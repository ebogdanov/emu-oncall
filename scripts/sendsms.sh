#!/bin/bash

text="You are invited to check and incident #222 with title \"ServiceUnreachable\", alert channel: AlertManager, alerts registered: 1, https://grafana-new.biz-systems.ru/a/grafana-oncall-app/alert-groups/I9BG3D69SI48H Your Grafana OnCall <3"
email="e.bogdanov@biz-systems.ru"

HOSTNAME="http://127.0.0.1:10080"

curl --data-urlencode "email=${email}" \
    --data-urlencode "message=${text}" \
    -v \
    "${HOSTNAME}/api/v1/send_sms/" | jq