package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	ConnectionString string     `json:"connection_string"`
	Databases        []Database `json:"databases"`
	Roles            []Role     `json:"roles"`
}

type Database struct {
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	Extensions []string `json:"extensions"'`
}

type Role struct {
	Name       string `json:"name""`
	Password   string `json:"password"`
	ParentRole string `json:"parent_role"`
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatal("Must supply one argument: the path to the config JSON file")
	}

	f, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	config := &Config{}
	err = json.Unmarshal(fBytes, config)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", config.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	for _, role := range config.Roles {
		fmt.Println("Configuring role: " + role.Name)
		rows, err := db.Query("SELECT rolname FROM pg_roles WHERE rolname = $1", role.Name)
		if err != nil {
			log.Fatal("error selecting rolname: ", role.Name, err)
		}
		defer rows.Close()
		if !rows.Next() {
			withClause := ""
			if role.ParentRole != "" {
				withClause = fmt.Sprintf(" WITH ROLE %s", role.ParentRole)
			}
			_, err = db.Exec(fmt.Sprintf("CREATE USER %s%s", role.Name, withClause))
			if err != nil {
				log.Fatal("error creating user: ", role.Name, err)
			}
		}
		_, err = db.Exec(fmt.Sprintf("ALTER USER %s WITH PASSWORD '%s'", role.Name, role.Password))
		if err != nil {
			log.Fatal("error setting user password: ", role.Name, err)
		}
	}

	for _, database := range config.Databases {
		fmt.Println("Configuring database: ", database.Name)
		rows, err := db.Query("SELECT datname FROM pg_database WHERE datname = $1", database.Name)
		if err != nil {
			log.Fatal("error selecting databases: ", database.Name, err)
		}
		defer rows.Close()
		if !rows.Next() {
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", database.Name, database.Owner))
			if err != nil {
				log.Fatal("error creating database: ", database.Name, err)
			}
			for _, ext := range database.Extensions {
				_, err = db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", ext))
				if err != nil {
					log.Fatal("error creating extension: ", ext, err)
				}
			}
		}
	}
}
