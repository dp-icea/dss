#!/usr/bin/env bash

set -eo pipefail

# Test Get SCD subscriptions.

subscription_id=$1
[[ -z "$subscription_id" ]] && { echo "Error: Subscription ID not provided"; exit 1; }

# Retrieve token from auth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8000/token?grant_type=client_credentials&scope=utm.strategic_coordination&intended_audience=localhost&issuer=localhost&sub=check_scd&apikey=brutm&aud=localhost" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  \
"http://localhost:8082/dss/v1/subscriptions/$subscription_id" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"
