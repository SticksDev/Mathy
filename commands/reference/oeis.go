package reference

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"mathy/commands"
	"mathy/utils"
	"mathy/logger"
)

type OEIS struct{}

func init() {
	commands.Register(&OEIS{})
}

func (o *OEIS) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("oeis", "Search the Online Encyclopedia of Integer Sequences").
		StringOption("query", "Sequence numbers (e.g., 1,1,2,3,5,8) or keywords", true).
		Build()
}

type oeisResult struct {
	Number  int      `json:"number"`
	Name    string   `json:"name"`
	Data    string   `json:"data"`
	Formula []string `json:"formula,omitempty"`
}

func (o *OEIS) HandleCommand(ctx *utils.Context) {
	query := ctx.Options()[0].StringValue()

	ctx.Defer(false)

	// OEIS API endpoint
	searchURL := fmt.Sprintf("https://oeis.org/search?fmt=json&q=%s", url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		logger.Error("OEIS request failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to connect to OEIS.",
			Color:       0xff0000,
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to read OEIS response.",
			Color:       0xff0000,
		})
		return
	}

	var results []oeisResult
	if err := json.Unmarshal(body, &results); err != nil {
		if strings.Contains(string(body), `"results":null`) {
			ctx.FollowupEmbed(&discordgo.MessageEmbed{
				Title:       "No Results",
				Description: fmt.Sprintf("No sequences found for: `%s`", query),
				Color:       0xffaa00,
			})
			return
		}
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to parse OEIS response.",
			Color:       0xff0000,
		})
		return
	}

	if len(results) == 0 {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("No sequences found for: `%s`", query),
			Color:       0xffaa00,
		})
		return
	}

	// Show top 3 results
	var embeds []*discordgo.MessageEmbed
	limit := 3
	if len(results) < limit {
		limit = len(results)
	}

	for i := 0; i < limit; i++ {
		seq := results[i]
		seqID := fmt.Sprintf("A%06d", seq.Number)
		seqURL := fmt.Sprintf("https://oeis.org/%s", seqID)

		// Truncate data to first 10 terms
		terms := strings.Split(seq.Data, ",")
		if len(terms) > 10 {
			terms = terms[:10]
		}
		dataPreview := strings.Join(terms, ", ") + ", ..."

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s: %s", seqID, truncate(seq.Name, 100)),
			URL:         seqURL,
			Description: fmt.Sprintf("**Sequence:** %s", dataPreview),
			Color:       0x3498db,
		}

		if len(seq.Formula) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Formula",
				Value: truncate(seq.Formula[0], 200),
			})
		}

		embeds = append(embeds, embed)
	}

	ctx.FollowupEmbed(embeds...)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
