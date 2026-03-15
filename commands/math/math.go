//go:build cgo

package math

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SticksDev/go-exprtk"
	"github.com/bwmarrin/discordgo"

	"mathy/commands"
	"mathy/utils"
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
	if err := parser.CompileExpression(); err != nil {
		errors := parser.CompileErrors()
		errorLines := err.Error()
		if len(errors) > 0 {
			errorLines = strings.Join(errors, "\n")
		}

		ctx.FollowupEmbed(&discordgo.MessageEmbed{
			Title:       "Expression Error",
			Description: fmt.Sprintf("Failed to evaluate:\n```\n%s\n```\n**Errors:**\n```\n%s\n```", expr, errorLines),
			Color:       utils.ColorError,
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
		Color: utils.ColorSuccess,
	})
}
