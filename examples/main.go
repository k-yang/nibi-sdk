package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/NibiruChain/nibiru/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	nibisdk "github.com/k-yang/nibi-sdk/src"
)

func main() {
	conf := sdk.GetConfig()
	if conf.GetBech32AccountAddrPrefix() != "nibi" {
		app.SetPrefixes(app.AccountAddressPrefix)
	}

	// Create Nibiru chain client
	slog.Info(
		"Creating chain client, connecting by GRPC",
		"chain-id", config.ChainId,
		"grpc-url", config.GrpcUrl,
		"grpc-insecure", config.GrpcInsecure,
	)
	chainId := "nibiru-localnet-0"
	grpcUrl := "localhost:9090"
	chainClient := nibisdk.NewChainClient(chainId, nibisdk.GetGRPCConnection(grpcUrl, true, 10*time.Second))
	account := chainClient.GetOrAddAccount("sender", config.Mnemonic)
	lastGasLimit := uint64(200000)

	// Transform records to WASM messages
	options := nibisdk.SendMsgOptions{
		Messages:     generateMessages(),
		SignerRecord: account,
		GasLimit:     lastGasLimit,
	}
	slog.Info(
		"Registering accounts",
		"count", len(options.Messages),
	)
	txResp, gasLimit := nibisdk.ExecuteWithGasRetries(chainClient, options, 0)
	if txResp != nil {
		slog.Info("Tx resp code: ", "code", txResp.Code)
	}
	lastGasLimit = gasLimit

	// Saving tx output
	// filename := fmt.Sprintf("resp-%05d-%05d-code-%d", i, lastIdx-1, txResp.Code)
	// filePath := filepath.Join(config.ResponsesOutputPath, filename)
	// slog.Info("Saving output", "path", filePath)
	// saveResponse(txResp, filePath)
}

// generateMessages - produces WASM message from the csv file record
func generateMessages(
	from sdk.AccAddress,
	to sdk.AccAddress,
) []sdk.Msg {
	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("unibi", sdk.NewInt(1e6))),
	})

	return msgs
}

// saveResponse - saves tx response to output
func saveResponse(
	txResp *sdk.TxResponse,
	filePath string,
) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	content := ""
	if txResp != nil {
		content = txResp.String()
	}
	if _, err := file.WriteString(content); err != nil {
		panic(err)
	}
}
