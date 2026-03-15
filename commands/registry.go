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

// MessageListener handles regular messages (not interactions).
type MessageListener interface {
	HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate)
}

var registry = make(map[string]Command)
var modalRegistry = make(map[string]ModalHandler)
var messageListeners []MessageListener
var registered []*discordgo.ApplicationCommand

func Register(cmd Command) {
	registry[cmd.Definition().Name] = cmd
}

func RegisterModal(customID string, handler ModalHandler) {
	modalRegistry[customID] = handler
}

func RegisterMessageListener(listener MessageListener) {
	messageListeners = append(messageListeners, listener)
}

// MessageRouter handles MessageCreate events, dispatching to all registered listeners.
func MessageRouter(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	for _, listener := range messageListeners {
		go listener.HandleMessage(s, m)
	}
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
	case discordgo.InteractionMessageComponent:
		utils.HandlePaginationInteraction(s, i)
	}
}

func RegisterForGuild(s *discordgo.Session, guildID string) {
	registered = bulkRegister(s, guildID)
}

func RegisterAll(s *discordgo.Session) {
	registered = bulkRegister(s, "")
}

func bulkRegister(s *discordgo.Session, guildID string) []*discordgo.ApplicationCommand {
	var defs []*discordgo.ApplicationCommand
	for _, cmd := range registry {
		defs = append(defs, cmd.Definition())
	}

	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, guildID, defs)
	if err != nil {
		logger.Error("Bulk command registration failed: %v", err)
		return nil
	}

	for _, cmd := range created {
		logger.Info("Registered command: /%s", cmd.Name)
	}
	return created
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
