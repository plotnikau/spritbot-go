# spritbot-go
A telegram bot for fuel price check written in golang: https://t.me/spritbot

Running serverless on [vercel](https://vercel.com)

### Environment settings
* ```TELEGRAM_BOT_TOKEN``` - telegram bot token, which you get from [BotFather](https://t.me/botfather)
  *  ```TELEGRAM_BOT_TOKEN_DEV``` can be set additionally on non-prod environment
* ```TK_API_KEY``` - [Tankerkoenig](https://creativecommons.tankerkoenig.de/) API key
* ```REDIS_URL``` - URL of your Redis instance, will be created automatically if you activate [vercel upstash integration](https://vercel.com/integrations/upstash)

### Telegram bot webhook setting
Use ```setwebhook.sh``` script with 2 arguments to set telegram webhook:
```setwebhook.sh <TELEGRAM_BOT_TOKEN> <HOOK_URL>```
