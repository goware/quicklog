package main

import (
	"fmt"
	"time"

	"github.com/goware/quicklog"
)

func main() {
	ql := quicklog.NewQuicklog()

	ql.Info("g1", "hi")
	ql.Info("g1", "test")
	ql.Warn("g1", "test")
	ql.Warn("g1", "test")
	time.Sleep(3 * time.Second)
	for _, l := range ql.Logs("g1") {
		fmt.Println(l.FormattedMessage("EST", true))
	}

	ql.Warn("g1", "test2")
	time.Sleep(10 * time.Second)
	fmt.Println(ql.Logs("g1")[1].TimeAgo())
	fmt.Println(ql.Logs("g1")[1].FormattedMessage("EST"))

	ql.Warn("g1", "test3")
	time.Sleep(100 * time.Second)
	fmt.Println(ql.Logs("g1")[2].TimeAgo())
	fmt.Println(ql.Logs("g1")[2].FormattedMessage("EST"))

}
