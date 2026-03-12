//go:build cgo

package math

import (
	"fmt"
	"mathy/commands"
	"mathy/utils"
	"strconv"
	"strings"
	"time"

	"github.com/SticksDev/go-exprtk"
	"github.com/bwmarrin/discordgo"
)

const (
	colorSuccess = 0x57F287 // green
	colorError   = 0xED4245 // red
)

type Math struct{}

func init() {
	commands.Register(&Math{})
}

func (m *Math) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("math", "Evaluate a math expression").
		StringOption("expression", "The math expression to evaluate", true).
		Build()
}

func (m *Math) HandleCommand(ctx *utils.Context) {
	startTime := time.Now()
	expr := ctx.Options()[0].StringValue()

	ctx.Defer(false)

	parser := exprtk.NewExprtk()
	defer parser.Delete()

	parser.SetExpression(expr)
	err := parser.CompileExpression()
	if err != nil {
		errors := parser.CompileErrors()
		var errorLines string
		if len(errors) > 0 {
			errorLines = strings.Join(errors, "\n")
		} else {
			errorLines = err.Error()
		}

		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Expression Error",
			Description: fmt.Sprintf("Failed to evaluate:\n```\n%s\n```\n**Errors:**\n```\n%s\n```", expr, errorLines),
			Color:       colorError,
		})
		return
	}

	result := parser.GetEvaluatedValue()
	formatted := strconv.FormatFloat(result, 'f', -1, 64)
	elapsed := time.Since(startTime)

	ctx.FollowupEmbed(&discordgo.MessageEmbed{
		Title: "Result",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Expression", Value: fmt.Sprintf("```\n%s\n```", expr), Inline: false},
			{Name: "Result", Value: fmt.Sprintf("`%s`", formatted), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Evaluated in %s", elapsed),
		},
		Color: colorSuccess,
	})
}
