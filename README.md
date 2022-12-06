## Usage

1. Get your TELEGRAM_BOT_TOKEN:

Create a Telegram bot by following the instructions in the Telegram Bot [documentation](https://core.telegram.org/bots#3-how-do-i-create-a-bot). This will require you to use the [BotFather](https://telegram.me/BotFather) bot to create a new bot and obtain a bot token.

2. Get your GPT3_API_KEY:

Create an OpenAI account by visiting the [OpenAI website](https://beta.openai.com/) and following the instructions on the sign-up page. Once you have an account, you can obtain your API key by visiting the API keys page in the OpenAI [documentation](https://beta.openai.com/docs/api-keys/overview).

3. Config your `.env` file:

Create a new file named .env in the root directory of your project, and add the following lines to it, replacing `xxxxxx` with your OpenAI API key and Telegram bot token:

```plain
GPT3_API_KEY="xxxxxx"
TELEGRAM_BOT_TOKEN="xxxxxx"
```

- Run the bot:

```bash
go run main.go
```
