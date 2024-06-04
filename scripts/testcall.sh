#!/bin/bash

HOSTNAME="emu-oncall-test-caller.test.biz-systems.ru"
# HOSTNAME="127.0.0.1:8089"

JSON='{
  "phone": "+79281972661",
  "description": "Description",
  "text": "Test"
}'

curl -X 'POST' \
  "http://$HOSTNAME/api/v1/call_to" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -v \
  -d '{
  "phone": "+79281972661",
  "description": "Description",
  "text": "Alert. Super text"
}'