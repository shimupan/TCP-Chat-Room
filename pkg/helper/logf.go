package helper

import "fmt"

func Logf(format string, args ...interface{}) {
	fmt.Printf("\r"+format+"> ", args...)
}
