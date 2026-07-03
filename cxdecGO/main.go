package main

import (
	"fmt"
	"os"

	"cxdecgo/internal/bootstrapdump"
	"cxdecgo/internal/unpackxp3"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "-d":
		if len(args) < 2 {
			return
		}
		_, err := bootstrapdump.Run(args[1])
		if err != nil {
			fmt.Println(err)
		}
	case "-x":
		if len(args) < 2 {
			return
		}
		err := unpackxp3.Run(args[1], false)
		if err != nil {
			fmt.Println(err)
		}
	case "-r":
		if len(args) < 2 {
			return
		}
		err := unpackxp3.Run(args[1], true)
		if err != nil {
			fmt.Println(err)
		}
	}
}
