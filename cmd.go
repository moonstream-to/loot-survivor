package main

import (
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func CreateRootCommand() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "survivor",
		Short: "Loot Survivor leaderboards by Moonstream",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	completionCmd := CreateCompletionCommand(rootCmd)
	versionCmd := CreateVersionCommand(rootCmd)
	starknetCmd := CreateStarknetCommand(rootCmd)
	rootCmd.AddCommand(completionCmd, versionCmd, starknetCmd)

	return rootCmd
}

func CreateCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts for survivor",
		Long: `Generate shell completion scripts for survivor.

The command for each shell will print a completion script to stdout. You can source this script to get
completions in your current shell session. You can add this script to the completion directory for your
shell to get completions for all future sessions.

For example, to activate bash completions in your current shell:
		$ . <(survivor completion bash)

To add survivor completions for all bash sessions:
		$ survivor completion bash > /etc/bash_completion.d/survivor_completions`,
	}

	bashCompletionCmd := &cobra.Command{
		Use:   "bash",
		Short: "bash completions for survivor",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(cmd.OutOrStdout())
		},
	}

	zshCompletionCmd := &cobra.Command{
		Use:   "zsh",
		Short: "zsh completions for survivor",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(cmd.OutOrStdout())
		},
	}

	fishCompletionCmd := &cobra.Command{
		Use:   "fish",
		Short: "fish completions for survivor",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}

	powershellCompletionCmd := &cobra.Command{
		Use:   "powershell",
		Short: "powershell completions for survivor",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
		},
	}

	completionCmd.AddCommand(bashCompletionCmd, zshCompletionCmd, fishCompletionCmd, powershellCompletionCmd)

	return completionCmd
}

func CreateVersionCommand(rootCmd *cobra.Command) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of survivor that you are currently using",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}

	return versionCmd
}

func CreateStarknetCommand(rootCmd *cobra.Command) *cobra.Command {
	var providerURL, RPCVersion string
	var timeout uint64

	starkCmd := &cobra.Command{
		Use:   "stark",
		Short: "Interact with your Starknet RPC provider",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	starkCmd.PersistentFlags().StringVarP(&providerURL, "provider", "p", os.Getenv("STARKNET_RPC_URL"), "The URL of your Starknet RPC provider (defaults to value of STARKNET_RPC_URL environment variable)")
	starkCmd.PersistentFlags().StringVarP(&RPCVersion, "rpcversion", "v", "2.0", "The version of the Starknet RPC protocol to use")
	starkCmd.PersistentFlags().Uint64VarP(&timeout, "timeout", "t", 0, "The timeout for requests to your Starknet RPC provider")

	blocknumberCmd := &cobra.Command{
		Use:   "blocknumber",
		Short: "Get the current block number on your Starknet RPC provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := Starknet{
				RPCVersion:  RPCVersion,
				ProviderURL: providerURL,
				Client:      &http.Client{Timeout: time.Duration(timeout) * time.Second},
			}
			blocknumber, err := provider.Blocknumber()
			if err != nil {
				return err
			}

			cmd.Println(blocknumber)
			return nil
		}}

	starkCmd.AddCommand(blocknumberCmd)

	return starkCmd
}
