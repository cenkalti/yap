package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/yap/task"
	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
)

// DefaultYapHome is the directory where yap keeps all task and configuration files.
const DefaultYapHome = "~/.yap"

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
		return task.SetHome(c.GlobalString("home"))
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
			Action:  cmdComplete,
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

func cmdList(c *cli.Context) {
	tasks, err := task.List()
	if err != nil {
		log.Fatal(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"ID", "Title"})
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
