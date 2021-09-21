#/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: setwebhook.sh BOT_API_TOKEN WEBHOOK_BASE_URL"
    exit 0
fi

BOT_API_TOKEN=$1
WEBHOOK_BASE_URL=$2

curl https://api.telegram.org/bot${BOT_API_TOKEN}/deleteWebhook
curl https://api.telegram.org/bot${BOT_API_TOKEN}/setWebhook?url=${WEBHOOK_BASE_URL}/api/telegram
