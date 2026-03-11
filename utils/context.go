package utils

import "github.com/bwmarrin/discordgo"

// Context. Passed to every command handler, contains the session and interaction details.
type Context struct {
	// The Discord session, used to send responses and followups.
	Session *discordgo.Session

	// The interaction that triggered the command. Contains all the details about the command invocation.
	Interaction *discordgo.InteractionCreate
}

type ResponseType int

const (
	ResponseMessage ResponseType = iota
	ResponseDeferred
)

// Response. Used to send replies to the user. Can be a normal message or a deferred response.
type Response struct {
	Type      ResponseType
	Content   string
	Embeds    []*discordgo.MessageEmbed
	Ephemeral bool
}

// Reply to the interaction with the specified response. Handles both normal and deferred responses, as well as ephemeral messages.
func (ctx *Context) Reply(r Response) error {
	responseType := discordgo.InteractionResponseChannelMessageWithSource
	if r.Type == ResponseDeferred {
		responseType = discordgo.InteractionResponseDeferredChannelMessageWithSource
	}

	var flags discordgo.MessageFlags
	if r.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Content: r.Content,
			Embeds:  r.Embeds,
			Flags:   flags,
		},
	})
}

func (ctx *Context) Defer(emphemeral bool) error {
	return ctx.Reply(Response{
		Type:      ResponseDeferred,
		Ephemeral: emphemeral,
	})
}

// Create a followup message to the interaction. This is used for sending additional messages after the initial response, especially for deferred responses.
func (ctx *Context) Followup(content string) error {
	_, err := ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	return err
}

// Create a followup message with embeds.
func (ctx *Context) FollowupEmbed(embeds ...*discordgo.MessageEmbed) error {
	_, err := ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Embeds: embeds,
	})
	return err
}

// Reply with a modal dialog.
func (ctx *Context) ReplyModal(customID, title string, components ...discordgo.MessageComponent) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   customID,
			Title:      title,
			Components: components,
		},
	})
}

// Get the options provided with the command invocation. This is a helper method to easily access the command arguments.
func (ctx *Context) Options() []*discordgo.ApplicationCommandInteractionDataOption {
	return ctx.Interaction.ApplicationCommandData().Options
}

// Get the modal submit data from the interaction.
func (ctx *Context) ModalData() discordgo.ModalSubmitInteractionData {
	return ctx.Interaction.ModalSubmitData()
}
