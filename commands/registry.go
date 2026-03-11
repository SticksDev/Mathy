package commands

import (
	"mathy/logger"
	"mathy/utils"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Definition() *discordgo.ApplicationCommand
	HandleCommand(ctx *utils.Context)
}

type ModalHandler interface {
	HandleModalSubmit(ctx *utils.Context)
}

var registry = make(map[string]Command)
var modalRegistry = make(map[string]ModalHandler)
var registered []*discordgo.ApplicationCommand

func Register(cmd Command) {
	registry[cmd.Definition().Name] = cmd
}

func RegisterModal(customID string, handler ModalHandler) {
	modalRegistry[customID] = handler
}

func Router(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &utils.Context{Session: s, Interaction: i}

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		name := i.ApplicationCommandData().Name
		if cmd, ok := registry[name]; ok {
			cmd.HandleCommand(ctx)
		}
	case discordgo.InteractionModalSubmit:
		customID := i.ModalSubmitData().CustomID
		if handler, ok := modalRegistry[customID]; ok {
			handler.HandleModalSubmit(ctx)
		}
	}
}

func RegisterForGuild(s *discordgo.Session, guildID string) {
	for _, cmd := range registry {
		created, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd.Definition())
		if err != nil {
			logger.Error("Cannot create command %q: %v", cmd.Definition().Name, err)
			continue
		}
		registered = append(registered, created)
		logger.Info("Registered command: /%s", created.Name)
	}
}

func RegisterAll(s *discordgo.Session) {
	// Register globally
	for _, cmd := range registry {
		created, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd.Definition())
		if err != nil {
			logger.Error("Cannot create command %q: %v", cmd.Definition().Name, err)
			continue
		}
		registered = append(registered, created)
		logger.Info("Registered command: /%s", created.Name)
	}
}

func RemoveAll(s *discordgo.Session, guildID string) {
	for _, cmd := range registered {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			logger.Error("Cannot delete command %q: %v", cmd.Name, err)
			continue
		}
		logger.Info("Removed command: /%s", cmd.Name)
	}
}
