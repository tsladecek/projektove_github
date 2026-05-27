package main

import "log"

func main() {
	if err := run(); err != nil {
		log.Fatalf("Received error when running code: %v", err)
	}
}

func run() error {
	return nil
}
