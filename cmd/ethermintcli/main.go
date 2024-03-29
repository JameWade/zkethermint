package main

import (
	"fmt"
	"github.com/cosmos/ethermint/zktx"
	"github.com/spf13/cobra"
	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
	"path/filepath"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	clientrpc "github.com/cosmos/cosmos-sdk/client/rpc"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/client"
	"github.com/cosmos/ethermint/codec"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/rpc"
	ethermint "github.com/cosmos/ethermint/types"
)

var (
	cdc = codec.MakeCodec(app.ModuleBasics)
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	tmamino.RegisterKeyType(ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName)
	tmamino.RegisterKeyType(ethsecp256k1.PrivKey{}, ethsecp256k1.PrivKeyName)

	keys.CryptoCdc = cdc
	clientkeys.KeysCdc = cdc

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	ethermint.SetBech32Prefixes(config)
	ethermint.SetBip44CoinType(config)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "ethermintcli",
		Short: "Command line interface for interacting with ethermintd",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return client.InitConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		clientrpc.StatusCommand(),
		sdkclient.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		client.ValidateChainID(
			rpc.ServeCmd(cdc),
		),
		flags.LineBreak,
		client.KeyCommands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// Add flags and prefix all env exposed with EM
	executor := cli.PrepareMainCmd(rootCmd, "EM", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		panic(fmt.Errorf("failed executing CLI command: %w", err))
	}
	//setFNfile()
}

func queryCmd(cdc *sdkcodec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		flags.LineBreak,
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *sdkcodec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		authcmd.GetDecodeCommand(cdc),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == auth.ModuleName || cmd.Use == bank.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)
	setFNfile()

	return txCmd
}

func setFNfile()  {
	DBdir  := os.ExpandEnv("$HOME/.ethermintcli")
	//DBdir, _ := filepath.Abs("$HOME/.ethermintcli")
	SNFilePath := filepath.Join(DBdir, "SN")
	SNfile, errOpenFile := os.OpenFile(SNFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if errOpenFile != nil {
		fmt.Println("OpenFile error: ", errOpenFile)
	}
	zktx.SNfile = SNfile
	////rd := bufio.NewReader(zktx.SNfile)
	////SSNBytesString2, errReading := rd.ReadString('\n')
	////
	////if errReading != nil {
	////	fmt.Println("Readiong string error: ", errReading)
	////}
	//var SSNBytesString2 string
	//var SNS zktx.SequenceS
	//if len(SSNBytesString2) != 0 {
	//	SSNBytesString := SSNBytesString2[0 : len(SSNBytesString2)-1]
	//	SNSbytes, errDecodeString := hex.DecodeString(SSNBytesString)
	//	if errDecodeString != nil {
	//		fmt.Println("Decode string  error: ", errDecodeString)
	//	}
	//	errDecodeBytes := rlp.DecodeBytes(SNSbytes, &SNS)
	//	if errDecodeBytes != nil {
	//		fmt.Println("Decode SNSbytes error: ", errDecodeBytes)
	//	}
	//	zktx.SequenceNumber = &SNS.Suquence1
	//	zktx.SequenceNumberAfter = &SNS.Suquence2
	//	zktx.SNS = SNS.SNS
	//	zktx.RandomReceiverPK = &ecdsa.PublicKey{crypto.S256(), SNS.PKBX, SNS.PKBY}
	//	zktx.Stage = SNS.Stage
	//}
	//zktx.SNfile.Close()
	//fmt.Println(zktx.SNfile)
}

