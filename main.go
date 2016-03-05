package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/satori/go.uuid"
)

var (
	yapHome           string
	tasksDir          string
	pendingTasksDir   string
	completedTasksDir string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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
	app.Before = func(c *cli.Context) (err error) {
		if yapHome, err = homedir.Expand(c.GlobalString("home")); err != nil {
			return
		}
		tasksDir = filepath.Join(yapHome, "tasks")
		if err = os.MkdirAll(tasksDir, 0700); err != nil {
			return
		}
		pendingTasksDir = filepath.Join(tasksDir, "pending")
		if err = os.MkdirAll(pendingTasksDir, 0700); err != nil {
			return
		}
		completedTasksDir = filepath.Join(tasksDir, "completed")
		if err = os.MkdirAll(completedTasksDir, 0700); err != nil {
			return
		}
		return
	}
	app.Action = cmdAdd // Default subcommand is "add".
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add new task",
			Action:  cmdAdd,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
			Action:  cmdList,
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

func cmdAdd(c *cli.Context) {
	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		return
	}
	id, err := NextTaskID(pendingTasksDir)
	if err != nil {
		log.Fatal(err)
	}
	t := PendingTask{
		ID: id,
		Task: Task{
			UUID:      uuid.NewV1(),
			Title:     strings.Join(c.Args(), " "),
			CreatedAt: time.Now(),
		},
	}
	if err = t.Task.Write(); err != nil {
		log.Fatal(err)
	}
	if err = t.Link(); err != nil {
		log.Fatal(err)
	}
}

func cmdList(c *cli.Context) {
	tasks, err := ListTasks()
	if err != nil {
		log.Fatal(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Title"})
	for _, v := range tasks {
		table.Append([]string{v.Title})
	}
	table.Render()
}
