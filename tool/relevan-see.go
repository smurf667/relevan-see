package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"relevan-see/config"
	"relevan-see/filters"
	"slices"
	"strings"

	"github.com/eiannone/keyboard"
)

func RootFolder() string {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	root := filepath.Dir(executable)
	for strings.Contains(root, ".git") {
		root = filepath.Join(root, "..")
	}
	return root
}

func Delta(root string) []filters.Modification {
	var buffer bytes.Buffer
	cmd := exec.Command("git", "diff-index", "--cached", "HEAD", "--")
	cmd.Dir = root
	cmd.Stdout = &buffer
	err := cmd.Run()
	if err != nil {
		slog.Warn("Cannot fetch Git modifications. Use --debug to see more.", slog.String("error", err.Error()))
		slog.Debug("Git", slog.String("output", buffer.String()))
		os.Exit(0)
	}

	scanner := bufio.NewScanner(strings.NewReader(buffer.String()))
	var result []filters.Modification
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 1 {
			var modification filters.Modification
			modification.OldHash = fields[2]
			modification.NewHash = fields[3]
			modification.Type = fields[4]
			modification.Name = fmt.Sprintf("./%s", fields[5])
			result = append(result, modification)
		}
	}
	return result
}

func main() {
	root := RootFolder()
	args := os.Args
	if slices.IndexFunc(args, func(arg string) bool { return arg == "--debug" }) > 0 {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if slices.IndexFunc(args, func(arg string) bool { return arg == "--dump-config" }) > 0 {
		config.DumpConfig(root)
		os.Exit(0)
	}
	if os.Getenv("CI") == `true` {
		slog.Debug("In a continuous integration setup, quick exit.")
		os.Exit(0)
	}
	if len(args) < 2 {
		slog.Error("Last argument must be the commit message file name.")
		fmt.Println()
		fmt.Print("Usage: relevan-see [--debug] [--dump-config] path-to-commit-message-file")
		os.Exit(1)
	}
	messageFile := args[len(args)-1]
	bytes, err := os.ReadFile(messageFile)
	if err != nil {
		slog.Error("Cannot read commit message file (must be last argument to tool).", slog.String("name", messageFile))
		os.Exit(1)
	}
	messageContent := string(bytes)
	cfg, err := config.Load(root)
	if err != nil {
		panic(err)
	}
	if strings.Contains(messageContent, cfg.Message) {
		slog.Debug("This commit already has a skip message")
		os.Exit(0)
	}

	slog.Debug("Running from", slog.String("folder", root))
	modifications := Delta(root)
	if len(modifications) == 0 {
		slog.Debug("No modifications found.")
		os.Exit(0)
	}

	all, err := filters.Init(root, cfg)
	if err != nil {
		panic(err)
	}

	// filter all changes (drop irrelevant changes from the list, if empty = no relevant change in commit)
	list := modifications
	for _, current := range all {
		list = current.Filter(list)
		if len(list) == 0 {
			messageContent := regexp.MustCompile(`(?sm)([^\r\n]*)(\r?\n.*)`).ReplaceAllString(messageContent, fmt.Sprintf("${1}%s${2}", cfg.Message))
			fmt.Println("\033[36mrelevan-see suspects no build-worthy changes!\033[0m")
			fmt.Println()
			fmt.Println("\033[32mChanges in commit:\033[0m")
			for _, entry := range modifications {
				fmt.Printf("  %s %s\n", entry.Type, entry.Name)
			}
			fmt.Println()
			fmt.Println("\033[32mSuggested message:\033[33m")
			fmt.Print(messageContent)
			fmt.Println("\033[0m")
			fmt.Printf("  [a]ccept\n  [c]ontinue as is\n  [q]uit\n\n  Enter your choice: ")
			if err := keyboard.Open(); err != nil {
				slog.Debug("No keyboard available")
				os.Exit(0)
			}
			defer func() {
				_ = keyboard.Close()
			}()

			var choice rune = ' '
			for choice == ' ' {
				char, _, err := keyboard.GetKey()
				if err != nil {
					os.Exit(1)
				}
				switch char {
				case 'a':
					outFile, err := os.OpenFile(messageFile, os.O_WRONLY, 0660)
					if err != nil {
						slog.Warn("Cannot write message file", slog.String("name", messageFile))
					}
					defer outFile.Close()
					outFile.Write([]byte(messageContent))
					os.Exit(0)
				case 'c':
					os.Exit(0)
				case 'q':
					os.Exit(1)
				}
			}
		} else {
			slog.Debug("Filter ends", slog.String("survivors", fmt.Sprintf("%v", list)))
		}
	}
	slog.Debug("Relevant", slog.String("changes", fmt.Sprintf("%v", list)))
	os.Exit(0)
}
