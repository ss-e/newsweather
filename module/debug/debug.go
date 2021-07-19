package debug

import (
	"fmt"
	"time"
)

func Output(module string, output string) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000") + " " + module + ": " + output)
	return
}
