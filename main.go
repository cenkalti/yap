package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cenkalti/yap/task"
	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/theckman/go-flock"
)

// DefaultYapHome is the directory where yap keeps all task and configuration files.
const DefaultYapHome = "~/.yap"

var instanceLock *flock.Flock

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
			Name:  "home",
			Value: DefaultYapHome,
			Usage: "home dir for yap",
		},
	}
	app.Before = func(c *cli.Context) error {
		home := c.GlobalString("home")
		var err error
		home, err = homedir.Expand(home)
		if err != nil {
			return err
		}
		err = task.SetHome(home)
		if err != nil {
			return err
		}
		lockPath := filepath.Join(home, ".lock")
		instanceLock = flock.NewFlock(lockPath)
		locked, err := instanceLock.TryLock()
		if err != nil {
			return err
		}
		if !locked {
			log.Fatal("another instance is running")
		}
		return nil
	}
	app.After = func(c *cli.Context) error {
		return os.Remove(instanceLock.Path())
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
	pt, err := task.Add(title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("id:", pt.LinkID)
}

func cmdListPending(c *cli.Context) {
	tasks, err := task.ListPending()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title")
	for _, v := range tasks {
		table.Append([]string{strconv.FormatUint(uint64(v.LinkID), 10), v.Title})
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
		table.Append([]string{strconv.FormatUint(uint64(v.LinkID), 10), v.Title})
	}
	table.Render()
}

func cmdComplete(c *cli.Context) {
	for _, arg := range c.Args() {
		id, err := strconv.ParseUint(arg, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		err = task.Complete(uint32(id))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cmdContinue(c *cli.Context) {
	for _, arg := range c.Args() {
		id, err := strconv.ParseUint(arg, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		err = task.Continue(uint32(id))
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
