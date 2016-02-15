package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
)

var yapHome string
var tasksDir string

type Task struct {
	ID    int64
	Title string
}

func NewTaskFromFile(dir, name string) (t Task, err error) {
	id := name[:len(name)-5]
	t.ID, err = strconv.ParseInt(id, 36, 64)
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, " ", 2)
		if len(parts) != 2 {
			err = errors.New("invalid task file")
			return
		}
		key, value := parts[0], parts[1]
		switch key {
		case "title":
			t.Title = value
		default:
			err = errors.New("invalid key")
			return
		}
	}
	err = scanner.Err()
	return
}

func (t Task) Line() string {
	return strconv.FormatInt(t.ID, 36) + " " + t.Title
}

func (t Task) WriteToFile(dir string) error {
	f, err := os.Create(filepath.Join(dir, strconv.FormatInt(t.ID, 36)) + ".task")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString("title " + t.Title + "\n")
	err = w.Flush()
	if err != nil {
		return err
	}
	return f.Close()
}

func ListTasks() ([]Task, error) {
	f, err := os.Open(tasksDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, name := range names {
		if !strings.HasSuffix(name, ".task") {
			continue
		}
		t, err := NewTaskFromFile(tasksDir, name)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func NextTaskID() (id int64, err error) {
	tasks, err := ListTasks()
	if err != nil {
		return
	}
	ids := make(map[int64]struct{})
	for _, task := range tasks {
		ids[task.ID] = struct{}{}
	}
	for id = int64(1); ; id++ {
		_, ok := ids[id]
		if !ok {
			return
		}
	}
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
	}
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add new task",
			Action: func(c *cli.Context) {
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
			},
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
