package main

import (
	"errors"
	"fmt"
	"gator/internal/config"
	"log"
	"os"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username is required for login")
	}
	username := cmd.args[0]
	if err := s.config.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %v", err)
	}
	fmt.Printf("User set to: %s\n", username)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	appState := &state{config: &cfg}

	cmdRegistry := &commands{}
	cmdRegistry.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("Error: Not enough arguments provided.")
		os.Exit(1)
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	cmd := command{name: cmdName, args: cmdArgs}

	if err := cmdRegistry.run(appState, cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// 	userName := "runner989"
// 	if err := cfg.SetUser(userName); err != nil {
// 		log.Fatalf("Error updating user in config: %v", err)
// 	}

// 	updatedCfg, err := config.Read()
// 	if err != nil {
// 		log.Fatalf("Error reading updated config: %v", err)
// 	}

// 	fmt.Printf("updated config: %+v\n", updatedCfg)
// }
