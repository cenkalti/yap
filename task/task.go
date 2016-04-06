package task

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/yap/datetime"
	"github.com/cenkalti/yap/kvfile"
	"github.com/satori/go.uuid"
)

const taskExt = ".task"

// Task is a file stored in tasks dir.
type Task struct {
	ID          uint16
	UUID        uuid.UUID
	Title       string             `key:"title"`
	CreatedAt   time.Time          `key:"created_at"`
	CompletedAt *time.Time         `key:"completed_at"`
	DueDate     *datetime.DateTime `key:"due_date"`
	WaitDate    *datetime.DateTime `key:"wait_date"`
}

func readFile(filename string) (t Task, err error) {
	defer func() {
		if err != nil {
			err = errors.New("cannot parse " + filename + ": " + err.Error())
		}
	}()

	base := filepath.Base(filename)
	t.UUID, err = uuid.FromString(base[:len(base)-len(taskExt)])
	if err != nil {
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer checkClose(f)

	err = kvfile.NewDecoder(f).Decode(&t)
	return
}

// write the task to file at <dirTasks>/<UUID>.task
func (t Task) write() error {
	path := filepath.Join(dirTasks, t.UUID.String()) + taskExt

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer checkClose(f)

	err = kvfile.NewEncoder(f).Encode(t)
	if err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

type timeCodec struct{}

func (c *timeCodec) Encode(i interface{}) string {
	return i.(time.Time).Format(time.RFC3339Nano)
}

func (c *timeCodec) Decode(s string) (interface{}, error) {
	return time.Parse(time.RFC3339Nano, s)
}

type dateTimeCodec struct{}

func (c *dateTimeCodec) Encode(i interface{}) string {
	return i.(datetime.DateTime).String()
}

func (c *dateTimeCodec) Decode(s string) (interface{}, error) {
	return datetime.Parse(s)
}

func init() {
	kvfile.RegisterCodec(new(time.Time), new(timeCodec))
	kvfile.RegisterCodec(new(datetime.DateTime), new(dateTimeCodec))
}
