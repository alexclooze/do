package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/alexflint/go-arg"
	"github.com/dittusch/go-shlex"
	. "github.com/logrusorgru/aurora"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var args struct {
	TaskName[] string `arg:"positional"`
}

type Dofile struct {
	Description string
	Tasks map[string]task
}

type task struct {
	Commands []string
	Tasks []string
	Output bool
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func parseCommand(command string) []string {
	parts, err := shlex.Split(strings.TrimSpace(command))
	if err != nil {
		log.Fatal(err)
	}

	return parts
}

func executeTask(doFile Dofile, taskName string) {
	if _, found := doFile.Tasks[taskName]; found {
		fmt.Println(Bold("-> Executing task\t"), Bold(Magenta(taskName)))

		for _, command := range doFile.Tasks[taskName].Commands {
			fmt.Println("  ", Bold(Yellow(taskName)), " ", command)

			tokens := parseCommand(command)
			cmdName := tokens[0]
			tokens = remove(tokens, 0)

			cmd := exec.Command(cmdName, tokens...)

			if _, err := os.Stat(cmdName); os.IsNotExist(err) {
				fmt.Println()
				fmt.Println(Bold(Red("Error: Command")), Bold(cmdName), Bold(Red("does not exist or is not in your $PATH!")))
				os.Exit(1)
			}

			if doFile.Tasks[taskName].Output == true {
				out, _ := cmd.CombinedOutput()

				//TODO: Identify if the executable does not exist
				//if err != nil {
				//	log.Fatalf("cmd.Run() failed with %s\n", err)
				//}
				fmt.Println()
				fmt.Println(Bold(Yellow("Output:")))
				fmt.Println(Yellow("--------------------------------------------------------------------------"))
				fmt.Printf(string(out))
				fmt.Println(Yellow("--------------------------------------------------------------------------"))
				fmt.Println()
			} else {
				if err := cmd.Run(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

		for _, task := range doFile.Tasks[taskName].Tasks {
			fmt.Println(Bold(Magenta("-> Executing subtask\t")), Bold(task))

			executeTask(doFile, task)
		}
	} else {
		fmt.Println(Bold(Red("Could not find task")), Bold(Yellow(taskName)), Bold(Red("aborting!")))
		os.Exit(-1);
	}
}

func main() {
	arg.MustParse(&args)

	fileContents, err := ioutil.ReadFile("./Dofile")
	if err != nil {
		log.Fatal(err)
	}

	var doFile Dofile
	if _, err := toml.Decode(string(fileContents), &doFile); err != nil {
		log.Fatal(err)
	}

	fmt.Println(Bold(Green(doFile.Description)))
	fmt.Println()

	for _, taskName := range args.TaskName {
		executeTask(doFile, taskName)
	}

	fmt.Println()
	fmt.Println(Bold(Green("Done executing all tasks for")), doFile.Description)
}
