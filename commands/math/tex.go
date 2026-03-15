package math

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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

var texPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\$\$(.+?)\$\$`),                    // $$...$$
	regexp.MustCompile(`\\\[(.+?)\\\]`),                     // \[...\]
	regexp.MustCompile(`(?:^|[^$])\$([^$]+?)\$(?:[^$]|$)`), // $...$
}

type Tex struct{}

func init() {
	tex := &Tex{}
	commands.Register(tex)
	commands.RegisterModal(modalID, tex)
	commands.RegisterMessageListener(tex)
}

// --- Slash command ---

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

// --- Modal submit ---

func (t *Tex) HandleModalSubmit(ctx *utils.Context) {
	latex := extractModalInput(ctx)
	if latex == "" {
		ctx.Reply(utils.Response{Content: "No LaTeX expression provided.", Ephemeral: true})
		return
	}

	ctx.Defer(false)
	startTime := time.Now()

	png, err := renderTeX(latex)
	if err != nil {
		logger.Error("Renderer request failed: %v", err)
		ctx.FollowupEmbed(texErrorEmbed(latex, err))
		return
	}

	ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{{
			Title: "LaTeX Render",
			Image: &discordgo.MessageEmbedImage{URL: "attachment://render.png"},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Rendered in %s", time.Since(startTime)),
			},
			Color: utils.ColorSuccess,
		}},
		Files: []*discordgo.File{{
			Name:        "render.png",
			ContentType: "image/png",
			Reader:      bytes.NewReader(png),
		}},
	})
}

// --- Message listener ---

func (t *Tex) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	blocks := extractTeX(m.Content)
	if len(blocks) == 0 {
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "⏳")

	var files []*discordgo.File
	hadError := false

	for i, block := range blocks {
		png, err := renderTeX(`\[` + block + `\]`)
		if err != nil {
			logger.Error("TeX listener render failed: %v", err)
			hadError = true
			continue
		}
		files = append(files, &discordgo.File{
			Name:        fmt.Sprintf("tex_%d.png", i+1),
			ContentType: "image/png",
			Reader:      bytes.NewReader(png),
		})
	}

	s.MessageReactionRemove(m.ChannelID, m.ID, "⏳", s.State.User.ID)

	if len(files) == 0 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
		return
	}

	if hadError {
		s.MessageReactionAdd(m.ChannelID, m.ID, "⚠️")
	}

	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Files:     files,
		Reference: m.Reference(),
	})
	if err != nil {
		logger.Error("TeX listener send failed: %v", err)
		s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
}

// --- Helpers ---

func extractModalInput(ctx *utils.Context) string {
	for _, row := range ctx.ModalData().Components {
		for _, comp := range row.(*discordgo.ActionsRow).Components {
			if input, ok := comp.(*discordgo.TextInput); ok && input.CustomID == inputID {
				return input.Value
			}
		}
	}
	return ""
}

func extractTeX(content string) []string {
	type region struct{ start, end int }

	var found []region
	var blocks []string

	for _, pat := range texPatterns {
		for _, loc := range pat.FindAllStringSubmatchIndex(content, -1) {
			start, end := loc[2], loc[3]
			overlaps := false
			for _, r := range found {
				if start < r.end && end > r.start {
					overlaps = true
					break
				}
			}
			if !overlaps && end > start {
				found = append(found, region{start, end})
				blocks = append(blocks, content[start:end])
			}
		}
	}

	return blocks
}

func renderTeX(latex string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"latex": latex})
	resp, err := http.Post(renderURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("renderer request failed: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("renderer returned %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func texErrorEmbed(latex string, err error) *discordgo.MessageEmbed {
	description := fmt.Sprintf("```\n%s\n```\n**Error:** %s", latex, err)

	// Try to parse renderer error details
	var errResp struct {
		Error   string   `json:"error"`
		Details []string `json:"details"`
	}
	if strings.Contains(err.Error(), "renderer returned") {
		// Error message contains the response body after the status code
		parts := strings.SplitN(err.Error(), ": ", 3)
		if len(parts) >= 3 {
			json.Unmarshal([]byte(parts[2]), &errResp)
		}
	}

	if errResp.Error != "" {
		description = fmt.Sprintf("```\n%s\n```\n**Error:** %s", latex, errResp.Error)
	}
	if len(errResp.Details) > 0 {
		description += fmt.Sprintf("\n```\n%s\n```", strings.Join(errResp.Details, "\n"))
	}

	return &discordgo.MessageEmbed{
		Title:       "LaTeX Error",
		Description: description,
		Color:       utils.ColorError,
	}
}
