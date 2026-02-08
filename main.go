package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/quanchobi/gator/internal/cli"
	"github.com/quanchobi/gator/internal/config"
	"github.com/quanchobi/gator/internal/database"
)

func main() {
	cmds := cli.Commands{
		Cmds: make(map[string]func(*cli.State, cli.Command) error),
	}

	functions := cli.GetFunctions()

	for name, handler := range functions {
		err := cmds.Register(name, handler)
		if err != nil {
			log.Fatal(err)
		}
	}

	args := os.Args[1:] // first argument will just be go, so we can ignore it
	if len(args) < 1 {
		log.Fatal(fmt.Errorf("at least one argument required"))
	}

	command := cli.Command{
		Name: args[0],  // function called
		Args: args[1:], // additional arguments
	}

	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	pdb, err := sql.Open("postgres", conf.DbURL)
	dbQueries := database.New(pdb)

	state := cli.State{
		Cfg: &conf,
		Db:  dbQueries,
	}

	err = cmds.Run(&state, command)
	if err != nil {
		log.Fatal(err)
	}
}
