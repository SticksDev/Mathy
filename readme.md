# Mathy

A simple, all in one, math bot for Discord. It can evaluate expressions, solve equations, and render TeX. Open source and self-hostable, so you can run it on your own server or locally.

We also have a [public instance](https://mathy.sticks.ovh) that you can invite to your server.

## Features

- Evaluate math expressions and solve equations
- Render TeX expressions to images
- Allows searching Wolfram Alpha and OEIS (Online Encyclopedia of Integer Sequences) for more complex queries
- Self-hostable and open source

## Self Hosting

To self host, you will need to have Docker and Docker Compose installed. This guide assumes you have a basic understanding of how to use these tools, as well as a local clone of the repository.

1. Copy the `.env.example` file to `.env` and fill in the required environment variables, such as your Discord bot token.
2. Run `docker-compose up -d` to start the bot and its dependencies (currently just a TeX rendering service).
3. Invite the bot to your discord server using the OAuth2 URL generated in the Discord Developer Portal.

## Wiki

This readme is just a brief overview of the bot and how to get it running. For more detailed information on how to use the bot, as well as troubleshooting and development guides, please refer to the [Wiki](https://github.com/MathyBot/Mathy/wiki) for the project.
