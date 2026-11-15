package main

import (
	"log"
	"os"

	"github.com/dunglas/frankenphp"
)

func main() {
	if len(os.Args) <= 1 {
		log.Println("Usage: testcli script.php")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--crash-in-c-thread":
		crashInCThread()
		os.Exit(0)
	case "--abort-in-c-thread":
		abortInCThread()
		os.Exit(0)
	case "--crash-in-go-thread":
		crashInGoThread()
		os.Exit(0)
	}

	os.Exit(frankenphp.ExecuteScriptCLI(os.Args[0], os.Args))
}
