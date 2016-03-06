package task

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func parseID(s string) (uint32, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	return uint32(i), err
}

func randomID() uint32 {
	return rand.Uint32()
}

func formatID(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}

// nextID returns the minimum available uint32 in dir.
func nextID(dir string) (id uint32, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	defer checkClose(f)
	names, err := f.Readdirnames(-1)
	if err != nil {
		return
	}
	ids := make(map[uint32]struct{})
	for _, name := range names {
		if !strings.HasSuffix(name, taskExt) {
			continue
		}
		strID := name[:len(name)-len(taskExt)]
		id, err = parseID(strID)
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
