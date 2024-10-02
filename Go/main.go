package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewSQLiteStorage()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		fmt.Printf("Nigga nekaj pa zdej ne deluje")
	}

	fmt.Printf("%+v\n", store)
	server := newAPIServer(":3000", store)
	server.Run()
	fmt.Println("Yeah buddy!")
}
