package main

import (
	"fmt"
	"os"

	"cxdecgo/internal/bootstrapdump"
)

func main() {
	args := os.Args[1:]

	switch args[0] {
	case "-d":
		{
			_, err := bootstrapdump.Run(args[1])
			if err != nil {
				fmt.Println(err)
			}
			break
		}

	}
}
