package main

import (
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
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
			Name:  "home",
			Value: DefaultYapHome,
			Usage: "home dir for yap",
		},
	}
	app.Before = func(c *cli.Context) error {
		return SetHome(c.GlobalString("home"))
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
			Name:    "pending",
			Aliases: []string{"l"},
			Usage:   "list pending tasks",
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
	// Since "add" is the default subcommand, we need to check if called with no args and show help.
	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		return
	}
	sid, err := NextTaskID(DirPendingTasks)
	if err != nil {
		log.Fatal(err)
	}
	t := PendingTask{
		TaskWithSmallID{
			SmallID: sid,
			Task: Task{
				ID:        rand.Uint32(),
				Title:     strings.Join(c.Args(), " "),
				CreatedAt: time.Now(),
			},
		},
	}
	if err = t.Task.Write(); err != nil {
		log.Fatal(err)
	}
	if err = t.Link(DirPendingTasks); err != nil {
		log.Fatal(err)
	}
}

func cmdList(c *cli.Context) {
	tasks, err := AllTasks()
	if err != nil {
		log.Fatal(err)
	}
	sort.Sort(ByCreatedAtDesc(tasks))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"ID", "Title"})
	for _, v := range tasks {
		table.Append([]string{strconv.FormatUint(uint64(v.ID), 10), v.Title})
	}
	table.Render()
}
