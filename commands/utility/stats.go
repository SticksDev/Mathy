package utility

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"

	"mathy/commands"
	"mathy/utils"
)

type Stats struct{}

var startTime = time.Now()

func init() {
	commands.Register(&Stats{})
}

func (s *Stats) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("stats", "Show bot statistics").Build()
}

func (s *Stats) HandleCommand(ctx *utils.Context) {
	uptime := time.Since(startTime)
	serverCount := len(ctx.Session.State.Guilds)

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memUsed := float64(mem.Alloc) / 1024 / 1024

	ctx.Reply(utils.Response{
		Embeds: []*discordgo.MessageEmbed{{
			Title: "Bot Statistics",
			Color: utils.ColorSuccess,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Servers", Value: fmt.Sprintf("%d", serverCount), Inline: true},
				{Name: "Uptime", Value: formatDuration(uptime), Inline: true},
				{Name: "Memory", Value: fmt.Sprintf("%.2f MB", memUsed), Inline: true},
				{Name: "Go Version", Value: runtime.Version(), Inline: true},
				{Name: "Goroutines", Value: fmt.Sprintf("%d", runtime.NumGoroutine()), Inline: true},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}},
	})
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
