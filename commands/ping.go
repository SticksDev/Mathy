package commands

import (
	"github.com/bwmarrin/discordgo"
	"mathy/utils"
)

type Ping struct{}

func init() {
	Register(&Ping{})
}

func (p *Ping) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("ping", "Pings the bot").Build()
}

func (p *Ping) HandleCommand(ctx *utils.Context) {
	ctx.Reply(utils.Response{Content: "Pong!"})
}
