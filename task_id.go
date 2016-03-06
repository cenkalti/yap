package main

import (
	"math/rand"
	"strconv"
)

type TaskID uint32

func ParseTaskID(s string) (TaskID, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	return TaskID(i), err
}

func NewRandomTaskID() TaskID {
	return TaskID(rand.Uint32())
}

func (i TaskID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}
