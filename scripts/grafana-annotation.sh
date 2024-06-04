#!/bin/sh

CI_GRAFANA_URL="http://localhost:3000/api/annotations"
CI_GRAFANA_DEPLOY_TOKEN="eyJrIjoiUnl2dk1WYTlBcW8xUmR3UzF4OFJLR1E1NWFmdWM4UngiLCJuIjoidGVzdC1hbm5vdGF0aW9uIiwiaWQiOjF9"

CI_PROJECT_NAME="ironmaiden"
CI_COMMIT_SHORT_SHA="abcdef123456"
GITLAB_USER_LOGIN="user-login"

curl -m 10 -s $CI_GRAFANA_URL \
  -H "Authorization: Bearer $CI_GRAFANA_DEPLOY_TOKEN" | jq

BODY=$(jq --null-input \
  --arg PROJECT_NAME "$CI_PROJECT_NAME" \
  --arg MSG "Deploy ${CI_PROJECT_NAME} commit: ${CI_COMMIT_SHORT_SHA}, by: ${GITLAB_USER_LOGIN}" \
  --arg EVENT_DATE $(date +%s000) \
  '{"time": $EVENT_DATE|tonumber, "text": $MSG, "tags":[$PROJECT_NAME, "deploy"]}')

echo $BODY

curl -m 10 -s -X POST $CI_GRAFANA_URL \
  -H "Authorization: Bearer $CI_GRAFANA_DEPLOY_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BODY" -v