#!/bin/sh

encoded_base64="YWRtaW46YWRtaW4="

curl -H "Authorization: Basic $encoded_base64" "http://grafana:3000/api/plugin-proxy/grafana-oncall-app/api/internal/v1/alertgroups?search=1"

curl 'https://grafana-new.biz-systems.ru/api/plugin-proxy/grafana-oncall-app/api/internal/v1/resolution_notes/' -X POST -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/109.0' -H 'Accept: application/json, text/plain, */*' -H 'Accept-Language: ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3' -H 'Accept-Encoding: gzip, deflate, br' -H 'Content-Type: application/json' -H 'Origin: https://grafana-new.biz-systems.ru' -H 'Connection: keep-alive' -H 'Referer: https://grafana-new.biz-systems.ru/a/grafana-oncall-app/incidents/I2AJT1115SFW2?perpage=25&start=1' -H 'Cookie: authelia_session=hLaN17f1QZJ6Nj4^czklj!nSJr#c-0fB; grafana_session=74fd75a19f2b11bf2a7602ec2b3277f7' -H 'Sec-Fetch-Dest: empty' -H 'Sec-Fetch-Mode: cors' -H 'Sec-Fetch-Site: same-origin' -H 'TE: trailers' --data-raw '{"alert_group":"I2AJT1115SFW2","text":"Acknowledged py phone +79281972661 by e.bogdanov"}'