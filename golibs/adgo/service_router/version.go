package service_router

import "fmt"

const (
	MAJOR_VERSION = 1
	MINOR_VERSION = 0
	REVISION      = 2
)

func Version() string {
	return fmt.Sprintf("%d.%d.%d", MAJOR_VERSION, MINOR_VERSION, REVISION)
}
