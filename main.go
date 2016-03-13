package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/cenkalti/yap/task"
	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
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
	app.Before = func(c *cli.Context) (err error) {
		fd, err := syscall.Open(task.Home, 0, 0)
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			log.Fatal(err)
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
			Subcommands: []cli.Command{
				{
					Name:    "pending",
					Aliases: []string{"p"},
					Usage:   "list pending tasks",
					Action:  cmdListPending,
				},
				{
					Name:    "completed",
					Aliases: []string{"c"},
					Usage:   "list completed tasks",
					Action:  cmdListCompleted,
				},
			},
		},
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task",
			Action:  cmdComplete,
		},
		{
			Name:    "continue",
			Aliases: []string{"con"},
			Usage:   "continue a completed task",
			Action:  cmdContinue,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdAdd(c *cli.Context) {
	// Since "add" is the default subcommand, we need to check if called with no args and show help.
	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		return
	}
	title := strings.Join(c.Args(), " ")
	id, err := task.Add(title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("id:", id)
}

func cmdListPending(c *cli.Context) {
	tasks, err := task.ListPending()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title")
	for _, v := range tasks {
		table.Append([]string{strconv.FormatUint(uint64(v.ID), 10), v.Title})
	}
	table.Render()
}

func cmdListCompleted(c *cli.Context) {
	tasks, err := task.ListCompleted()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title")
	for _, v := range tasks {
		table.Append([]string{strconv.FormatUint(uint64(v.ID), 10), v.Title})
	}
	table.Render()
}

func cmdComplete(c *cli.Context) {
	for _, arg := range c.Args() {
		id, err := strconv.ParseUint(arg, 10, 16)
		if err != nil {
			log.Fatal(err)
		}
		err = task.Complete(uint16(id))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cmdContinue(c *cli.Context) {
	for _, arg := range c.Args() {
		id, err := strconv.ParseUint(arg, 10, 16)
		if err != nil {
			log.Fatal(err)
		}
		err = task.Continue(uint16(id))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func newTable(fields ...string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoFormatHeaders(false)
	table.SetHeader(fields)
	return table
}
