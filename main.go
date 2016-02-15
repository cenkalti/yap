package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
)

var yapHome string
var tasksDir string

func main() {
	app := cli.NewApp()
	app.Name = "yap"
	app.Usage = "todo app"
	app.Authors = []cli.Author{
		{
			Name:  "Cenk Altı",
			Email: "cenkalti@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "home",
			Value:       "~/.yap",
			Usage:       "home dir for yap",
			Destination: &yapHome,
		},
	}
	app.Before = func(c *cli.Context) error {
		var err error
		yapHome, err = homedir.Expand(c.GlobalString("home"))
		if err != nil {
			return err
		}
		tasksDir = filepath.Join(yapHome, "tasks")
		return os.MkdirAll(tasksDir, 0700)
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			cli.ShowAppHelp(c)
			return
		}
		id, err := NextTaskID()
		if err != nil {
			log.Fatal(err)
		}
		t := Task{
			ID:    id,
			Title: strings.Join(c.Args(), " "),
		}
		err = t.WriteToFile(tasksDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add new task",
			Action:  app.Action,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
			Action: func(c *cli.Context) {
				tasks, err := ListTasks()
				if err != nil {
					log.Fatal(err)
				}
				for _, t := range tasks {
					fmt.Println(t.Line())
				}
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
