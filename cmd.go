package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
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
	versionCmd := CreateVersionCommand()
	starknetCmd := CreateStarknetCommand()
	abiCmd := CreateABICommand()
	findDeploymentBlockCmd := CreateFindDeploymentCmd()
	rootCmd.AddCommand(completionCmd, versionCmd, starknetCmd, abiCmd, findDeploymentBlockCmd)

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

func CreateVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of survivor that you are currently using",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}

	return versionCmd
}

func CreateABICommand() *cobra.Command {
	var abiFile string
	abiCmd := &cobra.Command{
		Use:   "abi",
		Short: "Interact with ABIs",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	abiCmd.PersistentFlags().StringVarP(&abiFile, "abi", "a", "", "The ABI file to interact with")

	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Lists the events in an ABI file",
		RunE: func(cmd *cobra.Command, args []string) error {
			infile, fileErr := os.Open(abiFile)
			if fileErr != nil {
				return fileErr
			}
			defer infile.Close()

			contents, readErr := io.ReadAll(infile)
			if readErr != nil {
				return readErr
			}

			var abi []map[string]interface{}
			unmarshalErr := json.Unmarshal(contents, &abi)
			if unmarshalErr != nil {
				return unmarshalErr
			}

			events, filterErr := Events(abi)
			if filterErr != nil {
				return filterErr
			}

			for _, event := range events {
				fmt.Printf("%s -- hash: %s\n", event.Name, event.Hash)
			}

			return nil
		},
	}

	abiCmd.AddCommand(eventsCmd)

	return abiCmd
}

func CreateStarknetCommand() *cobra.Command {
	var providerURL, contractAddress string
	var timeout, fromBlock, toBlock uint64
	var batchSize, coldInterval, hotInterval, hotThreshold, confirmations int

	starkCmd := &cobra.Command{
		Use:   "stark",
		Short: "Interact with your Starknet RPC provider",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	starkCmd.PersistentFlags().StringVarP(&providerURL, "provider", "p", os.Getenv("STARKNET_RPC_URL"), "The URL of your Starknet RPC provider (defaults to value of STARKNET_RPC_URL environment variable)")
	starkCmd.PersistentFlags().Uint64VarP(&timeout, "timeout", "t", 0, "The timeout for requests to your Starknet RPC provider")

	blockNumberCmd := &cobra.Command{
		Use:   "block-number",
		Short: "Get the current block number on your Starknet RPC provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, clientErr := rpc.NewClient(providerURL)
			if clientErr != nil {
				return clientErr
			}

			provider := rpc.NewProvider(client)

			ctx := context.Background()
			if timeout > 0 {
				ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Duration(timeout)*time.Second))
			}

			blockNumber, err := provider.BlockNumber(ctx)

			if err != nil {
				return err
			}

			cmd.Println(blockNumber)
			return nil
		}}

	chainIDCmd := &cobra.Command{
		Use:   "chain-id",
		Short: "Get the chain ID of the chain that your Starknet RPC provider is connected to",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, clientErr := rpc.NewClient(providerURL)
			if clientErr != nil {
				return clientErr
			}

			provider := rpc.NewProvider(client)

			ctx := context.Background()
			if timeout > 0 {
				ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Duration(timeout)*time.Second))
			}

			chainID, err := provider.ChainID(ctx)

			if err != nil {
				return err
			}

			cmd.Println(chainID)
			return nil
		}}

	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Crawl events from your Starknet RPC provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, clientErr := rpc.NewClient(providerURL)
			if clientErr != nil {
				return clientErr
			}

			provider := rpc.NewProvider(client)
			ctx := context.Background()

			eventsChan := make(chan CrawledEvent)

			go ContractEvents(ctx, provider, contractAddress, eventsChan, hotThreshold, time.Duration(hotInterval)*time.Millisecond, time.Duration(coldInterval)*time.Millisecond, fromBlock, confirmations, batchSize)

			for event := range eventsChan {
				fmt.Println(event)
			}

			return nil
		},
	}

	eventsCmd.Flags().StringVarP(&contractAddress, "contract", "c", "", "The address of the contract from which to crawl events (if not provided, no contract constraint will be specified)")
	eventsCmd.Flags().IntVarP(&batchSize, "batch-size", "N", 100, "The number of events to fetch per batch (defaults to 100)")
	eventsCmd.Flags().IntVar(&hotThreshold, "hot-threshold", 2, "Number of successive iterations which must return events before we consider the crawler hot")
	eventsCmd.Flags().IntVar(&hotInterval, "hot-interval", 100, "Milliseconds at which to poll the provider for updates on the contract while the crawl is hot")
	eventsCmd.Flags().IntVar(&coldInterval, "cold-interval", 10000, "Milliseconds at which to poll the provider for updates on the contract while the crawl is cold")
	eventsCmd.Flags().IntVar(&confirmations, "confirmations", 5, "Number of confirmations to wait for before considering a block canonical")
	eventsCmd.Flags().Uint64Var(&fromBlock, "from", 0, "The block number from which to start crawling")
	eventsCmd.Flags().Uint64Var(&toBlock, "to", 0, "The block number to which to crawl")

	starkCmd.AddCommand(blockNumberCmd, chainIDCmd, eventsCmd)

	return starkCmd
}

func CreateFindDeploymentCmd() *cobra.Command {
	var providerURL, contractAddress string

	findDeploymentCmd := &cobra.Command{
		Use:   "find-deployment-block",
		Short: "Discover the block number in which a contract was deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, clientErr := rpc.NewClient(providerURL)
			if clientErr != nil {
				return clientErr
			}
			provider := rpc.NewProvider(client)
			ctx := context.Background()

			if contractAddress == "" {
				return errors.New("you must provide a contract address using -c/--contract")
			}

			fieldAdditiveIdentity := fp.NewElement(0)
			if contractAddress[:2] == "0x" {
				contractAddress = contractAddress[2:]
			}
			decodedAddress, decodeErr := hex.DecodeString(contractAddress)
			if decodeErr != nil {
				return decodeErr
			}
			address := felt.NewFelt(&fieldAdditiveIdentity)
			address.SetBytes(decodedAddress)

			deploymentBlock, err := DeploymentBlock(ctx, provider, address)
			if err != nil {
				return err
			}

			cmd.Println(deploymentBlock)
			return nil
		},
	}

	findDeploymentCmd.Flags().StringVarP(&providerURL, "provider", "p", os.Getenv("STARKNET_RPC_URL"), "The URL of your Starknet RPC provider (defaults to value of STARKNET_RPC_URL environment variable)")
	findDeploymentCmd.Flags().StringVarP(&contractAddress, "contract", "c", "", "The address of the smart contract to find the deployment block for")

	return findDeploymentCmd
}
