package task

import (
	"os"
	"strconv"
	"strings"
)

func ParseID(s string) (uint16, error) {
	i, err := strconv.ParseUint(s, 10, 16)
	return uint16(i), err
}

func FormatID(i uint16) string {
	return strconv.FormatUint(uint64(i), 10)
}

// nextID returns the minimum available uint16 in dir.
func nextID(dir string) (id uint16, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	defer checkClose(f)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[uint16]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		strID := name[:len(name)-len(taskExt)]
		id, err = ParseID(strID)
		if err != nil {
			return
		}
		ids[id] = struct{}{}
	}
	for id = 1; ; id++ {
		_, ok := ids[id]
		if !ok {
			return
		}
	}
}
