package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"mathy/commands"
	_ "mathy/commands/math"
	"mathy/logger"
)

func main() {
	logger.Info("Starting Mathy...")
	godotenv.Load()

	token := os.Getenv("BOT_TOKEN")
	isDev := os.Getenv("DEV") == "true"
	guildId := os.Getenv("GUILD_ID")
	if token == "" {
		logger.Fatal("BOT_TOKEN is not set")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatal("Error creating session: %v", err)
	}

	session.AddHandler(commands.Router)

	err = session.Open()
	if err != nil {
		logger.Fatal("Error opening connection: %v", err)
	}
	defer session.Close()

	logger.Info("Connected to Discord as %s#%s", session.State.User.Username, session.State.User.Discriminator)

	if isDev && guildId != "" {
		logger.Info("Registering commands for guild %s... (development mode)", guildId)
		commands.RegisterForGuild(session, guildId)
	} else if isDev {
		logger.Warn("GUILD_ID is not set. Commands will not be registered in development mode.")
	} else {
		logger.Info("Starting in production mode. Registering global commands...")
		commands.RegisterAll(session)
	}

	logger.Info("Bot is running. Press Ctrl+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	logger.Info("Shutting down!")
	if isDev && guildId != "" {
		logger.Info("Removing commands from guild %s... (development mode)", guildId)
		commands.RemoveAll(session, guildId)
	}
}
