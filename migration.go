package gomigo

import (
	"fmt"
	"time"
)

type Migration struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	RegisteredAt time.Time `json:"registered_at"`
	Version      int       `json:"version"`
}

func ConstructName(name string) string {
	return fmt.Sprintf("%s_%s", time.Now().Format(Layout), name)
}
