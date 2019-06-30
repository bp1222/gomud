package game

type (
	Command interface {
		Act(client Client)
	}

	Echo struct {
		value string
	}

	Quit struct {
		value string
	}
)

func (c *Echo) Act(client Client) {
}

func (c *Quit) Act(client Client) {
	client.WriteString("You have left\n")
	client.Close()
}

var Commands = map[string]Command {
	"echo": &Echo{},
	"quit": &Quit{},
}
