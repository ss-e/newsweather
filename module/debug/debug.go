package debug

import (
	"fmt"
	"time"
)

func Output(module string, output string) {
	t := time.Now()
	fmt.Println("[" + t.Format("2006-01-02 15:04:05.000000") + "] " + module + ": " + output)
}
