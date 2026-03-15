package fun

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"mathy/commands"
	"mathy/utils"
)

type Roll struct{}

func init() {
	commands.Register(&Roll{})
}

func (r *Roll) Definition() *discordgo.ApplicationCommand {
	return utils.NewCommand("roll", "Roll dice (e.g., 2d6, 1d20+5, 3d8-2)").
		StringOption("dice", "Dice notation (e.g., 2d6, 1d20+5)", true).
		BoolOption("unsorted", "Keep results in roll order (don't sort)", false).
		Build()
}

var diceRegex = regexp.MustCompile(`^(\d+)?d(\d+)([+-]\d+)?$`)

func (r *Roll) HandleCommand(ctx *utils.Context) {
	opts := ctx.Options()
	dice := opts[0].StringValue()

	unsorted := false
	for _, opt := range opts {
		if opt.Name == "unsorted" {
			unsorted = opt.BoolValue()
		}
	}

	matches := diceRegex.FindStringSubmatch(strings.ToLower(dice))
	if matches == nil {
		ctx.Reply(utils.Response{
			Content:   "Invalid dice notation. Use format like `2d6`, `1d20+5`, or `3d8-2`.",
			Ephemeral: true,
		})
		return
	}

	numDice := 1
	if matches[1] != "" {
		numDice, _ = strconv.Atoi(matches[1])
	}
	sides, _ := strconv.Atoi(matches[2])
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	if numDice < 1 || numDice > 100 {
		ctx.Reply(utils.Response{Content: "Number of dice must be between 1 and 100.", Ephemeral: true})
		return
	}
	if sides < 2 || sides > 1000 {
		ctx.Reply(utils.Response{Content: "Number of sides must be between 2 and 1000.", Ephemeral: true})
		return
	}

	rolls := make([]int, numDice)
	sum := 0
	for i := range rolls {
		rolls[i] = rand.Intn(sides) + 1
		sum += rolls[i]
	}

	displayRolls := make([]int, len(rolls))
	copy(displayRolls, rolls)
	if !unsorted {
		sort.Sort(sort.Reverse(sort.IntSlice(displayRolls)))
	}

	total := sum + modifier

	rollStrs := make([]string, len(displayRolls))
	for i, roll := range displayRolls {
		rollStrs[i] = strconv.Itoa(roll)
	}

	var description string
	if numDice == 1 && modifier == 0 {
		description = fmt.Sprintf("**Result:** %d", total)
	} else {
		rollList := strings.Join(rollStrs, ", ")
		if modifier != 0 {
			sign := "+"
			if modifier < 0 {
				sign = ""
			}
			description = fmt.Sprintf("**Rolls:** [%s]\n**Sum:** %d %s%d = **%d**", rollList, sum, sign, modifier, total)
		} else {
			description = fmt.Sprintf("**Rolls:** [%s]\n**Total:** **%d**", rollList, total)
		}
	}

	sortNote := ""
	if !unsorted && numDice > 1 {
		sortNote = " (sorted)"
	}

	ctx.Reply(utils.Response{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       fmt.Sprintf("Rolling %s%s", dice, sortNote),
			Description: description,
			Color:       utils.ColorPurple,
		}},
	})
}
