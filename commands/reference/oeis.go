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
	"mathy/logger"
	"mathy/utils"
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

type oeisResponse struct {
	Results []oeisResult `json:"results"`
}

type oeisResult struct {
	Number  int      `json:"number"`
	Name    string   `json:"name"`
	Data    string   `json:"data"`
	Formula []string `json:"formula,omitempty"`
	Comment []string `json:"comment,omitempty"`
}

func (o *OEIS) HandleCommand(ctx *utils.Context) {
	query := ctx.Options()[0].StringValue()
	ctx.Defer(false)

	searchURL := fmt.Sprintf("https://oeis.org/search?fmt=json&q=%s&start=0", url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		logger.Error("OEIS request failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to connect to OEIS.",
			Color:       utils.ColorError,
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to read OEIS response.",
			Color:       utils.ColorError,
		})
		return
	}

	var oeisResp oeisResponse
	if err := json.Unmarshal(body, &oeisResp); err != nil || oeisResp.Results == nil {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("No sequences found for: `%s`", query),
			Color:       utils.ColorWarn,
		})
		return
	}

	results := oeisResp.Results
	if len(results) == 0 {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("No sequences found for: `%s`", query),
			Color:       utils.ColorWarn,
		})
		return
	}

	limit := min(len(results), 10)
	var pages []*discordgo.MessageEmbed

	for i := 0; i < limit; i++ {
		seq := results[i]
		seqID := fmt.Sprintf("A%06d", seq.Number)
		seqURL := fmt.Sprintf("https://oeis.org/%s", seqID)

		terms := strings.Split(seq.Data, ",")
		if len(terms) > 10 {
			terms = terms[:10]
		}
		dataPreview := strings.Join(terms, ", ") + ", ..."

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s: %s", seqID, utils.Truncate(seq.Name, 200)),
			URL:         seqURL,
			Description: fmt.Sprintf("**Sequence:** `%s`", dataPreview),
			Color:       utils.ColorInfo,
		}

		if len(seq.Formula) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Formula",
				Value: utils.Truncate(seq.Formula[0], 1024),
			})
		}

		if len(seq.Comment) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Comment",
				Value: utils.Truncate(seq.Comment[0], 1024),
			})
		}

		pages = append(pages, embed)
	}

	if len(pages) == 1 {
		ctx.FollowupEmbed(pages[0])
		return
	}

	paginator := utils.NewPaginatedEmbed(pages)
	paginator.Send(ctx)
}
