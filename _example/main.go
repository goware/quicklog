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
		fmt.Println(l.FormattedMessage("EST", false))
	}

	ql.Info("g1", "test")
	ql.Warn("g1", "test2")
	time.Sleep(10 * time.Second)
	for _, l := range ql.Logs("g1") {
		fmt.Println(l.FormattedMessage("EST", false))
	}

	ql.Warn("g1", "test3")
	time.Sleep(100 * time.Second)
	for _, l := range ql.Logs("g1") {
		fmt.Println(l.FormattedMessage("EST", false))
	}

}
