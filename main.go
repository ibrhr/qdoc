package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/docsource"
	"github.com/ibrhr/qdoc/internal/provider"
	"github.com/ibrhr/qdoc/internal/runner"
	"github.com/ibrhr/qdoc/internal/tui"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: loading config: %v\n", err)
	}

	// Pre-scan for flags
	var (
		useJSON  bool
		useNoTUI bool
		filtered []string
	)
	for _, a := range os.Args[1:] {
		switch a {
		case "--json":
			useJSON = true
		case "--no-tui":
			useNoTUI = true
		case "--version":
			fmt.Printf("qdoc %s (%s)\n", version, commit)
			return
		case "--help":
			printUsage()
			return
		default:
			filtered = append(filtered, a)
		}
	}

	if useJSON || useNoTUI {
		if len(filtered) < 2 {
			printUsage()
			os.Exit(1)
		}
		runHeadless(filtered[0], strings.Join(filtered[1:], " "), cfg, useJSON)
		return
	}

	if len(filtered) < 1 {
		if !provider.AnyKeyConfigured(cfg) {
			runFirstRunSetup(cfg)
			return
		}
		printUsage()
		return
	}

	cmd := filtered[0]
	args := filtered[1:]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("qdoc %s (%s)\n", version, commit)
		return
	case "status":
		printStatus(cfg)
		return
	case "sources", "ls":
		listSources()
		return
	case "providers":
		listProviders(cfg)
		return
	case "provider":
		runProviderSelect(cfg)
		return
	case "model":
		runModelSelect(cfg)
		return
	case "set":
		handleSet(args)
		return
	case "help", "--help", "-h":
		printUsage()
		return
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	sourceName := cmd
	query := strings.Join(args, " ")

	_, found := docsource.Find(sourceName)
	if !found {
		fmt.Fprintf(os.Stderr, "qdoc: unknown source %q\n\n", sourceName)
		listSources()
		os.Exit(1)
	}

	m := tui.NewQuery(sourceName, query, cfg)

	p := tea.NewProgram(&m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: %v\n", err)
		os.Exit(1)
	}
}

func runProviderSelect(cfg config.Config) {
	m := tui.NewProviderSelect(cfg)

	p := tea.NewProgram(&m)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: %v\n", err)
		os.Exit(1)
	}

	rm, ok := result.(*tui.Model)
	if ok && rm.SelectedProvider != "" {
		fmt.Printf("Provider set to %s.\n", rm.SelectedProvider)
	}
}

func runModelSelect(cfg config.Config) {
	providersWithKeys := tui.ProvidersWithKeys(cfg)

	if len(providersWithKeys) == 0 {
		fmt.Fprintf(os.Stderr, "qdoc: no providers with keys configured.\n")
		fmt.Fprintf(os.Stderr, "Run: qdoc set key <provider> <key>\n")
		os.Exit(1)
	}

	if len(providersWithKeys) == 1 {
		prov := providersWithKeys[0]
		m := tui.NewModelSelectSingle(cfg, prov)

		p := tea.NewProgram(&m)
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "qdoc: %v\n", err)
			os.Exit(1)
		}
		rm, ok := result.(*tui.Model)
		if ok {
			modelName := prov.DefaultModel
			if rm.Cfg.Models != nil && rm.Cfg.Models[prov.Name] != "" {
				modelName = rm.Cfg.Models[prov.Name]
			}
			fmt.Printf("Model for %s set to %s.\n", prov.Name, modelName)
		}
		return
	}

	m := tui.NewModelSelect(cfg)

	p := tea.NewProgram(&m)
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: %v\n", err)
		os.Exit(1)
	}
	rm, ok := result.(*tui.Model)
	if ok && rm.SelectedProvider != "" {
		modelName := rm.Cfg.Models[rm.SelectedProvider]
		if modelName == "" {
			prov, _ := provider.Find(rm.SelectedProvider)
			modelName = prov.DefaultModel
		}
		fmt.Printf("Model for %s set to %s.\n", rm.SelectedProvider, modelName)
	}
}

func handleSet(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "qdoc set <key|provider|model> <name> [value]\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  qdoc set provider openai\n")
		fmt.Fprintf(os.Stderr, "  qdoc set key openai sk-abc123...\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: loading config: %v\n", err)
	}

	switch args[0] {
	case "key":
		provName := args[1]
		if _, found := provider.Find(provName); !found {
			fmt.Fprintf(os.Stderr, "qdoc: unknown provider %q\n", provName)
			fmt.Fprintf(os.Stderr, "Available: openai, deepseek, opencode-zen, opencode-go\n")
			os.Exit(1)
		}

		var apiKey string
		if len(args) >= 3 {
			apiKey = args[2]
		} else {
			fmt.Printf("Enter API key for %s: ", provName)
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "qdoc: reading input: %v\n", err)
				os.Exit(1)
			}
			apiKey = string(b)
			fmt.Println()
		}

		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "qdoc: empty key provided\n")
			os.Exit(1)
		}

		cfg.Keys[provName] = apiKey
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "qdoc: saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("API key for %s saved.\n", provName)

	case "provider":
		provName := args[1]
		if _, found := provider.Find(provName); !found {
			fmt.Fprintf(os.Stderr, "qdoc: unknown provider %q\n", provName)
			fmt.Fprintf(os.Stderr, "Available: openai, deepseek, opencode-zen, opencode-go\n")
			os.Exit(1)
		}
		cfg.Provider = provName
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "qdoc: saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default provider set to %s.\n", provName)

	case "model":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "qdoc: missing model name\n")
			fmt.Fprintf(os.Stderr, "Usage: qdoc set model <provider> <model-name>\n")
			os.Exit(1)
		}
		provName := args[1]
		if _, found := provider.Find(provName); !found {
			fmt.Fprintf(os.Stderr, "qdoc: unknown provider %q\n", provName)
			fmt.Fprintf(os.Stderr, "Available: openai, deepseek, opencode-zen, opencode-go\n")
			os.Exit(1)
		}
		if cfg.Models == nil {
			cfg.Models = make(map[string]string)
		}
		cfg.Models[provName] = args[2]
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "qdoc: saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Model for %s set to %s.\n", provName, args[2])

	default:
		fmt.Fprintf(os.Stderr, "qdoc: unknown set command %q\n", args[0])
		fmt.Fprintf(os.Stderr, "Usage: qdoc set <key|provider|model>\n")
		os.Exit(1)
	}
}

func listProviders(cfg config.Config) {
	fmt.Println("Available LLM providers:")
	fmt.Println()
	for _, p := range provider.Providers {
		active := ""
		if cfg.Provider == p.Name {
			active = " (active)"
		}
		keyStatus := "no key"
		if cfg.Keys[p.Name] != "" {
			keyStatus = "key configured"
		}
		if os.Getenv(p.EnvKey) != "" {
			keyStatus = "env key set"
		}
		fmt.Printf("  %-14s %s%s\n", p.Name, keyStatus, active)
		fmt.Printf("    %s\n", p.Description)
		fmt.Printf("    env: %s\n", p.EnvKey)
		if m, ok := cfg.Models[p.Name]; ok && m != "" {
			fmt.Printf("    model: %s\n", m)
		} else {
			fmt.Printf("    model: %s (default)\n", p.DefaultModel)
		}
		fmt.Println()
	}
	fmt.Println("Commands:")
	fmt.Println("  qdoc provider              interactive provider picker")
	fmt.Println("  qdoc model                 interactive model picker")
	fmt.Println("  qdoc set provider <name>   switch default provider")
	fmt.Println("  qdoc set key <name> <key>  save an API key")
	fmt.Println("  qdoc set model <name> <m>  set model for provider")
}

func listSources() {
	fmt.Println("Available documentation sources:")
	fmt.Println()
	for _, s := range docsource.KnownSources {
		fmt.Printf("  %-12s %s\n", s.Name, s.BaseURL)
	}
	fmt.Println()
	fmt.Println("You can also point qdoc at a local directory of docs:")
	fmt.Println("  qdoc ./my-docs \"my question\"")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  qdoc provider                interactive provider picker")
	fmt.Println("  qdoc model                   interactive model picker")
	fmt.Println("  qdoc set provider <name>     switch default provider")
	fmt.Println("  qdoc set key <name> <key>    save an API key")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "qdoc %s - Query the Docs\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  qdoc <source> <query>        query a documentation source\n")
	fmt.Fprintf(os.Stderr, "  qdoc --json <source> <query> headless, JSON output\n")
	fmt.Fprintf(os.Stderr, "  qdoc --no-tui <source> <q>   headless, markdown output\n")
	fmt.Fprintf(os.Stderr, "  qdoc sources                 list documentation sources\n")
	fmt.Fprintf(os.Stderr, "  qdoc providers               list LLM providers and status\n")
	fmt.Fprintf(os.Stderr, "  qdoc status                  show current configuration\n")
	fmt.Fprintf(os.Stderr, "  qdoc provider                interactive provider selection\n")
	fmt.Fprintf(os.Stderr, "  qdoc model                   interactive model selection\n")
	fmt.Fprintf(os.Stderr, "  qdoc set <key|provider> ...  configure providers\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  qdoc go \"generics tutorial\"\n")
	fmt.Fprintf(os.Stderr, "  qdoc ./my-docs \"deployment guide\"\n")
	fmt.Fprintf(os.Stderr, "  qdoc --json go \"generics\"    agent-friendly output\n")
	fmt.Fprintf(os.Stderr, "  qdoc provider                pick a provider interactively\n")
	fmt.Fprintf(os.Stderr, "  qdoc set key openai          enter API key interactively\n")
}

func runFirstRunSetup(cfg config.Config) {
	fmt.Println("Welcome to qdoc! Let's get you set up.")
	fmt.Println()

	m := tui.NewSetup(cfg)

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: %v\n", err)
		os.Exit(1)
	}
}

func runHeadless(sourceName, query string, cfg config.Config, jsonOut bool) {
	result := runner.Run(sourceName, query, cfg)
	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "qdoc: %v\n", result.Err)
		os.Exit(1)
	}

	if jsonOut {
		out, _ := json.MarshalIndent(struct {
			Answer string         `json:"answer"`
			Source string         `json:"source"`
			Steps  []runner.Step `json:"steps"`
		}{result.Answer, result.Source, result.Steps}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Print(result.Answer)
	}
}

func printStatus(cfg config.Config) {
	provName := cfg.Provider
	if provName == "" {
		provName = "(not set)"
	}

	prov, found := provider.Find(provName)
	keyStatus := "no key"
	if found {
		if cfg.Keys[prov.Name] != "" {
			keyStatus = "key configured"
		} else if os.Getenv(prov.EnvKey) != "" {
			keyStatus = "env key set"
		}
	}

	modelName := "(default)"
	if cfg.Models != nil && cfg.Models[provName] != "" {
		modelName = cfg.Models[provName]
	} else if found {
		modelName = prov.DefaultModel + " (default)"
	}

	fmt.Println("qdoc configuration:")
	fmt.Println()
	fmt.Printf("  Provider:  %s (%s)\n", provName, keyStatus)
	fmt.Printf("  Model:     %s\n", modelName)
	fmt.Println()

	if found {
		fmt.Printf("  Base URL:  %s\n", prov.BaseURL)
		fmt.Printf("  Env var:   %s\n", prov.EnvKey)
	}
	fmt.Println()
	fmt.Println("Change with: qdoc provider  |  qdoc model  |  qdoc set key <provider>")
}