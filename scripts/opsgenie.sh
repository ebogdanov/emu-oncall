#!/bin/bash

curl -X GET 'https://api.opsgenie.com/v2/users' --header 'Authorization: GenieKey f902e1bb-c3b0-4d1d-b6ff-94c97d6e8c81' | jq ".data[] | {id: .id[0:8] | ascii_upcase , user: .username, name: .fullName, role: .role.name | ascii_downcase } "