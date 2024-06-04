#!/bin/bash

HOSTNAME="http://127.0.0.1:10080"
curl -v -X POST "${HOSTNAME}/api/v1/integrations/"