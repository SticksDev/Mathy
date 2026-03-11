package math

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"mathy/commands"
	"mathy/logger"
	"mathy/utils"
)

const (
	modalID   = "tex_modal"
	inputID   = "tex_input"
	renderURL = "http://renderer:8080/render"
)

type Tex struct{}

func init() {
	tex := &Tex{}
	commands.Register(tex)
	commands.RegisterModal(modalID, tex)
}

func (t *Tex) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("tex", "Render a LaTeX expression").Build()
}

func (t *Tex) HandleCommand(ctx *utils.Context) {
	ctx.ReplyModal(modalID, "LaTeX Expression",
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    inputID,
					Label:       "LaTeX Expression",
					Style:       discordgo.TextInputParagraph,
					Placeholder: `e.g. \frac{a}{b} + \sqrt{c}`,
					Required:    true,
					MinLength:   1,
					MaxLength:   2000,
				},
			},
		},
	)
}

func (t *Tex) HandleModalSubmit(ctx *utils.Context) {
	data := ctx.ModalData()

	var latex string
	for _, row := range data.Components {
		for _, comp := range row.(*discordgo.ActionsRow).Components {
			if input, ok := comp.(*discordgo.TextInput); ok && input.CustomID == inputID {
				latex = input.Value
			}
		}
	}

	if latex == "" {
		ctx.Reply(utils.Response{
			Content:   "No LaTeX expression provided.",
			Ephemeral: true,
		})
		return
	}

	ctx.Defer(false)
	startTime := time.Now()

	body, _ := json.Marshal(map[string]string{"latex": latex})
	resp, err := http.Post(renderURL, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Error("Renderer request failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Render Error",
			Description: "Could not connect to the renderer service.",
			Color:       colorError,
		})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
			Log   string `json:"log"`
		}
		json.Unmarshal(respBody, &errResp)

		description := fmt.Sprintf("```\n%s\n```", latex)
		if errResp.Error != "" {
			description += fmt.Sprintf("\n**Error:** %s", errResp.Error)
		}
		if errResp.Log != "" {
			// Trim log to last 500 chars to avoid embed limits
			log := errResp.Log
			if len(log) > 500 {
				log = log[len(log)-500:]
			}
			description += fmt.Sprintf("\n**Log:**\n```\n%s\n```", log)
		}

		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "LaTeX Error",
			Description: description,
			Color:       colorError,
		})
		return
	}

	elapsed := time.Since(startTime)

	ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "LaTeX Render",
				Image: &discordgo.MessageEmbedImage{
					URL: "attachment://render.svg",
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Rendered in %s", elapsed),
				},
				Color: colorSuccess,
			},
		},
		Files: []*discordgo.File{
			{
				Name:        "render.svg",
				ContentType: "image/svg+xml",
				Reader:      bytes.NewReader(respBody),
			},
		},
	})
}
