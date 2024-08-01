package nibisdk

import (
	"log/slog"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExecuteWithGasRetries sends message to a chain, retries if out of gas
func ExecuteWithGasRetries(
	chainClient ChainClient,
	options SendMsgOptions,
	attempt int,
) (*sdk.TxResponse, uint64) {
	txResp, err := chainClient.SendMsg(options)
	if err != nil {
		text := err.Error()
		if strings.Contains(text, "out of gas") {
			if attempt >= 5 {
				slog.Error(
					"Execution failed due to out of gas",
					"gas_limit", options.GasLimit,
					"attempt", attempt,
					"reason", text,
				)
				panic(err)
			}
			options.GasLimit *= 2
			slog.Warn(
				"Retrying due to out of gas",
				"gas_limit", options.GasLimit,
				"attempt", attempt+1,
			)

			return ExecuteWithGasRetries(chainClient, options, attempt+1)
		}

		panic(err)
	}

	if txResp.Code != 0 {
		text := txResp.RawLog
		if strings.Contains(text, "out of gas") {
			if attempt >= 5 {
				slog.Error(
					"Execution failed due to out of gas",
					"gas_limit", options.GasLimit,
					"attempt", attempt,
					"reason", text,
				)
				panic(text)
			}
			options.GasLimit *= 2
			slog.Warn(
				"Retrying due to out of gas",
				"gas_limit", options.GasLimit,
				"attempt", attempt+1,
			)

			return ExecuteWithGasRetries(chainClient, options, attempt+1)
		}
		panic(text)
	}

	return txResp, options.GasLimit
}
