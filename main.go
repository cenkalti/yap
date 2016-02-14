package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "yap"
	app.Usage = "todo app"
	app.Action = func(c *cli.Context) {
	}

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add new task",
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task",
			Action: func(c *cli.Context) {
			},
		},
	}

	app.Run(os.Args)
}
