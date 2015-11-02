package main

import (
	"os"
	"strings"
	"time"
)
import "fmt"

func main() {
	for i := 0; true; i++ {
		now := time.Now()
		timeString := strings.Replace(now.Format("2006-01-02 15:04:05.000"), ".", ",", 1)
		fmt.Printf("%s - DEBUG - loggertest - test logger message %d", timeString, i)
		fmt.Println()
	}
	os.Exit(0)
}
