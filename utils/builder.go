package utils

import "github.com/bwmarrin/discordgo"

// The command builder. Provides a fluent interface for constructing application commands.
type CommandBuilder struct {
	// The underlying command being built. This is what will eventually be registered with Discord.
	cmd *discordgo.ApplicationCommand
}

// Create a new command builder with the specified name and description.
func NewCommand(name, description string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &discordgo.ApplicationCommand{
			Name:        name,
			Description: description,
		},
	}
}

// Create a number option for the command.
func (b *CommandBuilder) NumberOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionNumber,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Create a string option for the command.
func (b *CommandBuilder) StringOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Create an integer option for the command.
func (b *CommandBuilder) IntOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Create a boolean option for the command.
func (b *CommandBuilder) BoolOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Create a user option for the command.
func (b *CommandBuilder) UserOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Create a channel option for the command.
func (b *CommandBuilder) ChannelOption(name, description string, required bool) *CommandBuilder {
	b.cmd.Options = append(b.cmd.Options, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionChannel,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// Finalize the command and return the underlying ApplicationCommand struct.
func (b *CommandBuilder) Build() *discordgo.ApplicationCommand {
	return b.cmd
}
