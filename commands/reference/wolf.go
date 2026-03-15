package reference

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"

	"mathy/commands"
	"mathy/logger"
	"mathy/utils"
)

type Wolf struct{}

func init() {
	commands.Register(&Wolf{})
}

func (w *Wolf) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("wolf", "Query Wolfram|Alpha").
		StringOption("query", "Your question or calculation", true).
		Build()
}

type wolframResponse struct {
	XMLName xml.Name     `xml:"queryresult"`
	Success bool         `xml:"success,attr"`
	Error   bool         `xml:"error,attr"`
	Pods    []wolframPod `xml:"pod"`
}

type wolframPod struct {
	Title   string          `xml:"title,attr"`
	Primary bool            `xml:"primary,attr"`
	SubPods []wolframSubPod `xml:"subpod"`
}

type wolframSubPod struct {
	Title     string `xml:"title,attr"`
	PlainText string `xml:"plaintext"`
}

func (w *Wolf) HandleCommand(ctx *utils.Context) {
	appID := os.Getenv("WOLFRAM_APP_ID")
	if appID == "" {
		ctx.Reply(utils.Response{
			Content:   "Wolfram|Alpha is not configured. Missing `WOLFRAM_APP_ID`.",
			Ephemeral: true,
		})
		return
	}

	query := ctx.Options()[0].StringValue()
	ctx.Defer(false)

	apiURL := fmt.Sprintf(
		"https://api.wolframalpha.com/v2/query?appid=%s&input=%s&format=plaintext",
		appID, url.QueryEscape(query),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		logger.Error("Wolfram Alpha request failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to connect to Wolfram|Alpha.",
			Color:       utils.ColorError,
		})
		return
	}
	defer resp.Body.Close()

	var result wolframResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("Wolfram Alpha XML decode failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to parse Wolfram|Alpha response.",
			Color:       utils.ColorError,
		})
		return
	}

	if !result.Success || len(result.Pods) == 0 {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("Wolfram|Alpha couldn't interpret: `%s`", query),
			Color:       utils.ColorWarn,
		})
		return
	}

	queryURL := fmt.Sprintf("https://www.wolframalpha.com/input?i=%s", url.QueryEscape(query))

	// Filter pods with text content, skip input interpretation
	var pods []wolframPod
	for _, pod := range result.Pods {
		if strings.Contains(strings.ToLower(pod.Title), "input") {
			continue
		}
		for _, sp := range pod.SubPods {
			if sp.PlainText != "" {
				pods = append(pods, pod)
				break
			}
		}
	}

	if len(pods) == 0 {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("Wolfram|Alpha returned no text results for: `%s`", query),
			Color:       utils.ColorWarn,
		})
		return
	}

	limit := min(len(pods), 10)
	var pages []*discordgo.MessageEmbed

	for i := 0; i < limit; i++ {
		pod := pods[i]

		var texts []string
		for _, sp := range pod.SubPods {
			if sp.PlainText != "" {
				texts = append(texts, sp.PlainText)
			}
		}
		content := strings.Join(texts, "\n")

		pages = append(pages, &discordgo.MessageEmbed{
			Title:       pod.Title,
			URL:         queryURL,
			Description: fmt.Sprintf("```\n%s\n```", utils.Truncate(content, 2000)),
			Color:       utils.ColorOrange,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Powered by Wolfram|Alpha",
			},
		})
	}

	if len(pages) == 1 {
		ctx.FollowupEmbed(pages[0])
		return
	}

	paginator := utils.NewPaginatedEmbed(pages)
	paginator.Send(ctx)
}
