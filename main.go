package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/cenkalti/yap/datetime"
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
	app.HideHelp = true
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "due",
					Usage: "due date. complete task before this date",
				},
				cli.StringFlag{
					Name:  "wait",
					Usage: "wait date. task will be hidden until this date",
				},
			},
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
				{
					Name:    "waiting",
					Aliases: []string{"w"},
					Usage:   "list waiting tasks",
					Action:  cmdListWaiting,
				},
			},
		},
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete pending task",
			Action:  cmdComplete,
		},
		{
			Name:    "continue",
			Aliases: []string{"con"},
			Usage:   "continue completed task",
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
	dueDate := parseDateTime(c.String("due"))
	waitDate := parseDateTime(c.String("wait"))
	id, err := task.Add(title, dueDate, waitDate)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("id:", id)
}

func parseDateTime(s string) *datetime.DateTime {
	if s == "" {
		return nil
	}
	dt, err := datetime.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
	return &dt
}

func formatDateTime(dt *datetime.DateTime) string {
	if dt == nil {
		return ""
	}
	return dt.String()
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return datetime.New(*t).String()
}

func cmdListPending(c *cli.Context) {
	tasks, err := task.ListPending()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title", "Due date")
	for _, t := range tasks {
		table.Append([]string{
			task.FormatID(t.ID),
			t.Title,
			formatDateTime(t.DueDate),
		})
	}
	table.Render()
}

func cmdListCompleted(c *cli.Context) {
	tasks, err := task.ListCompleted()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title", "Completed at")
	for _, t := range tasks {
		table.Append([]string{
			task.FormatID(t.ID),
			t.Title,
			formatDate(t.CompletedAt),
		})
	}
	table.Render()
}

func cmdListWaiting(c *cli.Context) {
	tasks, err := task.ListWaiting()
	if err != nil {
		log.Fatal(err)
	}
	table := newTable("ID", "Title", "Wait date", "Due date")
	for _, t := range tasks {
		table.Append([]string{
			task.FormatID(t.ID),
			t.Title,
			formatDateTime(t.WaitDate),
			formatDateTime(t.DueDate),
		})
	}
	table.Render()
}

func cmdComplete(c *cli.Context) {
	ids, err := parseIDs(c.Args())
	if err != nil {
		log.Fatal(err)
	}
	err = task.Complete(ids)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdContinue(c *cli.Context) {
	ids, err := parseIDs(c.Args())
	if err != nil {
		log.Fatal(err)
	}
	err = task.Continue(ids)
	if err != nil {
		log.Fatal(err)
	}
}

func parseIDs(args []string) ([]uint16, error) {
	ids := make([]uint16, 0, len(args))
	for _, arg := range args {
		id, err := task.ParseID(arg)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func newTable(fields ...string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetAutoFormatHeaders(false)
	table.SetHeader(fields)
	return table
}
