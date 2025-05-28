package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to nixai! Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Exiting nixai. Goodbye!")
			break
		}

		handleCommand(input)
	}
}

// handleCommand processes user commands entered in interactive mode.
func handleCommand(command string) {
	// TODO: Implement command handling logic
	fmt.Println("You entered:", command)
}
