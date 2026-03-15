package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// PaginatedEmbed holds a set of embeds that can be navigated with buttons.
type PaginatedEmbed struct {
	Pages   []*discordgo.MessageEmbed
	current int
	done    chan struct{}
}

var (
	paginators   = make(map[string]*PaginatedEmbed)
	paginatorsMu sync.Mutex
)

// NewPaginatedEmbed creates a paginator from a slice of embeds.
// It auto-adds page numbers to footers.
func NewPaginatedEmbed(pages []*discordgo.MessageEmbed) *PaginatedEmbed {
	for i, page := range pages {
		footer := fmt.Sprintf("Page %d/%d", i+1, len(pages))
		if page.Footer != nil && page.Footer.Text != "" {
			footer = page.Footer.Text + " • " + footer
		}
		page.Footer = &discordgo.MessageEmbedFooter{Text: footer}
	}
	return &PaginatedEmbed{Pages: pages, current: 0, done: make(chan struct{})}
}

func (p *PaginatedEmbed) buttons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "page_first",
					Emoji:    &discordgo.ComponentEmoji{Name: "⏮"},
					Style:    discordgo.SecondaryButton,
					Disabled: p.current == 0,
				},
				discordgo.Button{
					CustomID: "page_prev",
					Emoji:    &discordgo.ComponentEmoji{Name: "◀"},
					Style:    discordgo.SecondaryButton,
					Disabled: p.current == 0,
				},
				discordgo.Button{
					CustomID: "page_next",
					Emoji:    &discordgo.ComponentEmoji{Name: "▶"},
					Style:    discordgo.SecondaryButton,
					Disabled: p.current == len(p.Pages)-1,
				},
				discordgo.Button{
					CustomID: "page_last",
					Emoji:    &discordgo.ComponentEmoji{Name: "⏭"},
					Style:    discordgo.SecondaryButton,
					Disabled: p.current == len(p.Pages)-1,
				},
				discordgo.Button{
					CustomID: "page_delete",
					Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
}

// Send the paginated embed as a followup message and registers it for interaction.
func (p *PaginatedEmbed) Send(ctx *Context) error {
	msg, err := ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{p.Pages[0]},
		Components: p.buttons(),
	})
	if err != nil {
		return err
	}

	paginatorsMu.Lock()
	paginators[msg.ID] = p
	paginatorsMu.Unlock()

	// Auto-expire after 5 minutes
	go func() {
		select {
		case <-time.After(5 * time.Minute):
		case <-p.done:
		}
		paginatorsMu.Lock()
		delete(paginators, msg.ID)
		paginatorsMu.Unlock()
	}()

	return nil
}

// HandlePaginationInteraction handles button clicks for paginated embeds.
// Returns true if the interaction was handled.
func HandlePaginationInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Type != discordgo.InteractionMessageComponent {
		return false
	}

	customID := i.MessageComponentData().CustomID
	switch customID {
	case "page_first", "page_prev", "page_next", "page_last", "page_delete":
	default:
		return false
	}

	paginatorsMu.Lock()
	p, ok := paginators[i.Message.ID]
	paginatorsMu.Unlock()

	if !ok {
		return false
	}

	if customID == "page_delete" {
		close(p.done)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Dismissed.",
				Embeds:     []*discordgo.MessageEmbed{},
				Components: []discordgo.MessageComponent{},
			},
		})
		s.FollowupMessageDelete(i.Interaction, i.Message.ID)
		return true
	}

	switch customID {
	case "page_first":
		p.current = 0
	case "page_prev":
		if p.current > 0 {
			p.current--
		}
	case "page_next":
		if p.current < len(p.Pages)-1 {
			p.current++
		}
	case "page_last":
		p.current = len(p.Pages) - 1
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{p.Pages[p.current]},
			Components: p.buttons(),
		},
	})

	return true
}
