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

// Wolfram Alpha API response structures
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

	// Query Wolfram Alpha API
	apiURL := fmt.Sprintf(
		"https://api.wolframalpha.com/v2/query?appid=%s&input=%s&format=plaintext",
		appID,
		url.QueryEscape(query),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		logger.Error("Wolfram Alpha request failed: %v", err)
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Error",
			Description: "Failed to connect to Wolfram|Alpha.",
			Color:       0xff0000,
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
			Color:       0xff0000,
		})
		return
	}

	if !result.Success || len(result.Pods) == 0 {
		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "No Results",
			Description: fmt.Sprintf("Wolfram|Alpha couldn't interpret: `%s`", query),
			Color:       0xffaa00,
		})
		return
	}

	// Build embed with relevant pods
	embed := &discordgo.MessageEmbed{
		Title: truncate(query, 100),
		URL:   fmt.Sprintf("https://www.wolframalpha.com/input?i=%s", url.QueryEscape(query)),
		Color: 0xff6600,
	}

	// Find primary result and other interesting pods
	var primaryText string
	fieldsAdded := 0
	maxFields := 6

	for _, pod := range result.Pods {
		if fieldsAdded >= maxFields {
			break
		}

		// Skip input interpretation pod
		if strings.Contains(strings.ToLower(pod.Title), "input") {
			continue
		}

		// Get text from subpods
		var texts []string
		for _, subpod := range pod.SubPods {
			if subpod.PlainText != "" {
				texts = append(texts, subpod.PlainText)
			}
		}

		if len(texts) == 0 {
			continue
		}

		podText := strings.Join(texts, "\n")

		// If this is the primary pod, use it as description
		if pod.Primary && primaryText == "" {
			primaryText = podText
			continue
		}

		// Add as field
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  pod.Title,
			Value: truncate(podText, 1024),
		})
		fieldsAdded++
	}

	if primaryText != "" {
		embed.Description = truncate(primaryText, 2048)
	} else if len(embed.Fields) == 0 {
		// If no primary and no fields, use first pod content
		for _, pod := range result.Pods {
			for _, subpod := range pod.SubPods {
				if subpod.PlainText != "" {
					embed.Description = truncate(subpod.PlainText, 2048)
					break
				}
			}
			if embed.Description != "" {
				break
			}
		}
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Powered by Wolfram|Alpha",
	}

	ctx.FollowupEmbed(embed)
}
