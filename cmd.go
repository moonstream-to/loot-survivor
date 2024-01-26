package main

import (
	"bufio"
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
	leaderboardsCmd := CreateLeaderboardsCmd()
	reparseCmd := CreateParseCommand()
	rootCmd.AddCommand(completionCmd, versionCmd, starknetCmd, abiCmd, findDeploymentBlockCmd, leaderboardsCmd, reparseCmd)

	// By default, cobra Command objects write to stderr. We have to forcibly set them to output to
	// stdout.
	rootCmd.SetOut(os.Stdout)

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if providerURL == "" {
				providerURLFromEnv := os.Getenv("STARKNET_RPC_URL")
				if providerURLFromEnv == "" {
					return errors.New("you must provide a provider URL using -p/--provider or set the STARKNET_RPC_URL environment variable")
				}
				providerURL = providerURLFromEnv
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	starkCmd.PersistentFlags().StringVarP(&providerURL, "provider", "p", "", "The URL of your Starknet RPC provider (defaults to value of STARKNET_RPC_URL environment variable)")
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

			eventsChan := make(chan RawEvent)

			// If "fromBlock" is not specified, find the block at which the contract was deployed and
			// use that instead.
			if fromBlock == 0 {
				addressFelt, parseAddressErr := FeltFromHexString(contractAddress)
				if parseAddressErr != nil {
					return parseAddressErr
				}
				deploymentBlock, fromBlockErr := DeploymentBlock(ctx, provider, addressFelt)
				if fromBlockErr != nil {
					return fromBlockErr
				}
				fromBlock = deploymentBlock
			}

			go ContractEvents(ctx, provider, contractAddress, eventsChan, hotThreshold, time.Duration(hotInterval)*time.Millisecond, time.Duration(coldInterval)*time.Millisecond, fromBlock, toBlock, confirmations, batchSize)

			for event := range eventsChan {
				unparsedEvent := ParsedEvent{Name: EVENT_UNKNOWN, Event: event}
				serializedEvent, marshalErr := json.Marshal(unparsedEvent)
				if marshalErr != nil {
					cmd.ErrOrStderr().Write([]byte(marshalErr.Error()))
				}
				cmd.Println(string(serializedEvent))
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
	eventsCmd.Flags().Uint64Var(&toBlock, "to", 0, "The block number to which to crawl (set to 0 for continuous crawl)")

	starkCmd.AddCommand(blockNumberCmd, chainIDCmd, eventsCmd)

	return starkCmd
}

func CreateFindDeploymentCmd() *cobra.Command {
	var providerURL, contractAddress string

	findDeploymentCmd := &cobra.Command{
		Use:   "find-deployment-block",
		Short: "Discover the block number in which a contract was deployed",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if providerURL == "" {
				providerURLFromEnv := os.Getenv("STARKNET_RPC_URL")
				if providerURLFromEnv == "" {
					return errors.New("you must provide a provider URL using -p/--provider or set the STARKNET_RPC_URL environment variable")
				}
				providerURL = providerURLFromEnv
			}
			return nil
		},
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

	findDeploymentCmd.Flags().StringVarP(&providerURL, "provider", "p", "", "The URL of your Starknet RPC provider (defaults to value of STARKNET_RPC_URL environment variable)")
	findDeploymentCmd.Flags().StringVarP(&contractAddress, "contract", "c", "", "The address of the smart contract to find the deployment block for")

	return findDeploymentCmd
}

func CreateLeaderboardsCmd() *cobra.Command {
	var infile, outfile, leaderboardID, accessToken string
	var push bool

	leaderboardsCmd := &cobra.Command{
		Use:   "leaderboards",
		Short: "Generates Loot Survivor leaderboards and can push them to the Moonstream Leaderboards API",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if push {
				if leaderboardID == "" {
					leaderboardIDFromEnv := os.Getenv("MOONSTREAM_LEADERBOARD_ID")
					if leaderboardIDFromEnv == "" {
						return errors.New("when pushing, you must provide a leaderboard ID using -l/--leaderboard-id or set the MOONSTREAM_LEADERBOARD_ID environment variable")
					}
					leaderboardID = leaderboardIDFromEnv
				}
				if accessToken == "" {
					accessTokenFromEnv := os.Getenv("MOONSTREAM_ACCESS_TOKEN")
					if accessTokenFromEnv == "" {
						return errors.New("when pushing, you must provide an access token using -t/--access-token or set the MOONSTREAM_ACCESS_TOKEN environment variable")
					}
					accessToken = accessTokenFromEnv
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	leaderboardsCmd.PersistentFlags().StringVarP(&infile, "infile", "i", "", "File containing crawled events from which to build the leaderboard (as produced by the \"loot-survivor stark events\" command, defaults to stdin)")
	leaderboardsCmd.PersistentFlags().StringVarP(&outfile, "outfile", "o", "", "File to write leaderboard to (defaults to stdout)")
	leaderboardsCmd.PersistentFlags().BoolVar(&push, "push", false, "Set this option to pus the leaderboard to the Moonstream Leaderboard API")
	leaderboardsCmd.PersistentFlags().StringVarP(&leaderboardID, "leaderboard-id", "l", "", "Leaderboard ID for the Moonstream Leaderboard (look up or generate at https://moonstream.to, defaults to value of MOONSTREAM_LEADERBOARD_ID environment variable)")
	leaderboardsCmd.PersistentFlags().StringVarP(&accessToken, "access-token", "t", "", "Access token for Moonstream API (get from https://moonstream.to, defaults to value of MOONSTREAM_ACCESS_TOKEN environment variable)")

	totalCmd := &cobra.Command{
		Use:   "total",
		Short: "Leaderboard of all player events in Loot Survivor",
		Long: `Leaderboard of all player events in Loot Survivor

NOTE: This is a leaderboard of adventurers, not their owners.

This leaderboard awards a number of points to each event that an adventurer could be subject to. From
these points, it calculates a total Loot Survivor score for each adventurer. The leaderboard also reports
the individual event scores for each adventurer in the "points_data" field.

The leaderboard also lists the active owner for each adventurer, defined as the account that last used
the adventurer in a game session.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ifp := os.Stdin
			var infileErr error
			if infile != "" && infile != "-" {
				ifp, infileErr = os.Open(infile)
				if infileErr != nil {
					return infileErr
				}
				defer ifp.Close()
			}

			ofp := os.Stdout
			var outfileErr error
			if outfile != "" {
				ofp, outfileErr = os.Create(outfile)
				if outfileErr != nil {
					return outfileErr
				}
				defer ofp.Close()
			}

			leaderboard, leaderboardErr := LootSurvivorLeaderboard(ifp)
			if leaderboardErr != nil {
				return leaderboardErr
			}

			outputEncoder := json.NewEncoder(ofp)
			outputEncoder.Encode(leaderboard)

			if push {
				pushErr := Push(leaderboardID, accessToken, leaderboard, true)
				if pushErr != nil {
					return pushErr
				}
			}
			return nil
		},
	}

	leaderboardsCmd.AddCommand(totalCmd)

	return leaderboardsCmd
}

func CreateParseCommand() *cobra.Command {
	var infile, outfile string

	parseCmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse a file (as produced by the \"stark events\" command) to process previously unknown events",
		RunE: func(cmd *cobra.Command, args []string) error {
			ifp := os.Stdin
			var infileErr error
			if infile != "" && infile != "-" {
				ifp, infileErr = os.Open(infile)
				if infileErr != nil {
					return infileErr
				}
				defer ifp.Close()
			}

			ofp := os.Stdout
			var outfileErr error
			if outfile != "" {
				ofp, outfileErr = os.Create(outfile)
				if outfileErr != nil {
					return outfileErr
				}
				defer ofp.Close()
			}

			parser, newParserErr := NewEventParser()
			if newParserErr != nil {
				return newParserErr
			}

			newline := []byte("\n")

			scanner := bufio.NewScanner(ifp)
			for scanner.Scan() {
				var partialEvent PartialEvent
				line := scanner.Text()
				json.Unmarshal([]byte(line), &partialEvent)

				passThrough := true

				if partialEvent.Name == EVENT_UNKNOWN {
					var event RawEvent
					json.Unmarshal(partialEvent.Event, &event)
					parsedEvent, parseErr := parser.Parse(event)
					if parseErr == nil {
						passThrough = false

						parsedEventBytes, marshalErr := json.Marshal(parsedEvent)
						if marshalErr != nil {
							return marshalErr
						}

						_, writeErr := ofp.Write(parsedEventBytes)
						if writeErr != nil {
							return writeErr
						}
						_, writeErr = ofp.Write(newline)
						if writeErr != nil {
							return writeErr
						}
					}
				}

				if passThrough {
					partialEventBytes, marshalErr := json.Marshal(partialEvent)
					if marshalErr != nil {
						return marshalErr
					}

					_, writeErr := ofp.Write(partialEventBytes)
					if writeErr != nil {
						return writeErr
					}
					_, writeErr = ofp.Write(newline)
					if writeErr != nil {
						return writeErr
					}
				}
			}

			return nil
		},
	}

	parseCmd.Flags().StringVarP(&infile, "infile", "i", "", "File containing crawled events from which to build the leaderboard (as produced by the \"loot-survivor stark events\" command, defaults to stdin)")
	parseCmd.Flags().StringVarP(&outfile, "outfile", "o", "", "File to write reparsed events to (defaults to stdout)")

	return parseCmd
}
