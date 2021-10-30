package main

import (
	"github.com/CoolGoose/hugosync/cmd/hugosync"
	"log"
	"os"
)

func main() {
	err := hugosync.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
