package main

import (
	"flag"
	"fmt"
	"log"
)

// seed the database with an account for testing purposes
func seedAccount(store Storage, fname, pw string) *Account {
	acc, err := NewAccount(fname, pw)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("neww account => ", acc.Number)

	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "saibot", "saibotsan")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.init(); err != nil {
		log.Fatal(err)
	}
	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)

	}

	server := NewAPIServer(":3000", store)
	server.Run()
}
