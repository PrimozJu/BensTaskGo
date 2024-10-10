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
		fmt.Printf("Cannot initialize DB")
	}

	fmt.Printf("%+v\n", store)
	server := newAPIServer(":3000", store)
	server.Run()
	fmt.Println("All good!")

}
