package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/quanchobi/gator/internal/config"
	"github.com/quanchobi/gator/internal/database"
)

type State struct {
	Cfg *config.Config
	Db  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmds map[string]func(*State, Command) error
}

func GetFunctions() map[string]func(*State, Command) error {
	return map[string]func(*State, Command) error{
		"login":    HandlerLogin,
		"register": HandlerRegister,
	}
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		fmt.Println("login expects one argument, the username.")
		os.Exit(1)
	}
	username := cmd.Args[0]
	if username == "" {
		fmt.Println("login expects one argument, the username.")
		os.Exit(1)
	}
	// check if user is in database
	user, err := s.Db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}
	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%s has logged in successfully.\n", username)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		fmt.Println("register expects one argument, the username")
		os.Exit(1)
	}
	username := cmd.Args[0]
	if username == "" {
		fmt.Println("register expects one argument, the username.")
		os.Exit(1)
	}

	user, err := s.Db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)
	if err != nil {
		return err
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("%s has registered successfully.\n", user.Name)
	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	fn, ok := c.Cmds[cmd.Name]
	if !ok {
		fmt.Printf("command %s not found", cmd.Name)
	}
	err := fn(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) error {
	_, exists := c.Cmds[name]
	if exists {
		return fmt.Errorf("attempted re-register of function signature %T", f) // prolly shouldnt happen
	}
	c.Cmds[name] = f
	return nil
}
