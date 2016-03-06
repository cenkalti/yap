package task

import (
	"io"
	"log"
)

func checkClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
