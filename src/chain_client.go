package nibisdk

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/NibiruChain/nibiru/app"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
)

type ChainClient struct {
	chainId  string
	keyring  keyring.Keyring
	encCfg   app.EncodingConfig
	grpcConn *grpc.ClientConn
	txClient txTypes.ServiceClient
}

type SendMsgOptions struct {
	Messages     []sdk.Msg
	SignerRecord *keyring.Record
	GasLimit     uint64
}

func NewChainClient(chainId string, conn *grpc.ClientConn) ChainClient {
	encCfg := app.MakeEncodingConfig()

	client := ChainClient{
		chainId:  chainId,
		keyring:  keyring.NewInMemory(encCfg.Marshaler),
		encCfg:   encCfg,
		grpcConn: conn,
		txClient: txTypes.NewServiceClient(conn),
	}

	return client
}

func (chainClient *ChainClient) SendMsg(options SendMsgOptions) (*sdk.TxResponse, error) {
	signerAddress, err := options.SignerRecord.GetAddress()
	if err != nil {
		return nil, err
	}

	txBuilder := chainClient.encCfg.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(options.Messages...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(options.GasLimit)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("unibi", int64(options.GasLimit/40))))
	txBuilder.SetFeePayer(signerAddress)

	accNum, seqNum := chainClient.GetAccountNumbers(signerAddress)
	txFactory := tx.Factory{}.
		WithChainID(chainClient.chainId).
		WithAccountNumber(accNum).
		WithSequence(seqNum).
		WithKeybase(chainClient.keyring).
		WithTxConfig(chainClient.encCfg.TxConfig)

	err = tx.Sign(txFactory, options.SignerRecord.Name, txBuilder, true)
	if err != nil {
		return nil, err
	}

	// Generated Protobuf-encoded bytes.
	txBytes, err := chainClient.encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	// Broadcast message
	ctx := context.Background()
	grpcRes, err := chainClient.txClient.BroadcastTx(
		ctx,
		&txTypes.BroadcastTxRequest{
			Mode:    txTypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		slog.Error("Error broadcasting transaction", "err", err)
		return nil, err
	}
	if grpcRes.TxResponse.Code != 0 {
		slog.Error("Error broadcasting transaction", "code", grpcRes.TxResponse.Code, "log", grpcRes.TxResponse.RawLog)
		return nil, errors.New(grpcRes.TxResponse.RawLog)
	}

	// Wait while transaction is committed
	txHash := grpcRes.TxResponse.TxHash
	slog.Info("Transaction sent. Waiting for response", "hash", txHash)

	timeout := time.NewTimer(time.Minute)
	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-tick.C:
			resp, _ := chainClient.txClient.GetTx(ctx, &txTypes.GetTxRequest{Hash: txHash})
			if resp != nil && resp.TxResponse != nil {
				timeout.Stop()
				tick.Stop()
				return resp.TxResponse, nil
			}
		case <-timeout.C:
			timeout.Stop()
			tick.Stop()
			return nil, errors.New("create transaction timeout error")
		}
	}
}

type AccountNumbers struct {
	Number   uint64
	Sequence uint64
}

// GetAccountNumbers returns account number and sequence number for an address
func (chainClient *ChainClient) GetAccountNumbers(address sdk.AccAddress) (accNum uint64, seqNum uint64) {
	authClient := authTypes.NewQueryClient(chainClient.grpcConn)
	resp, err := authClient.Account(
		context.Background(),
		&authTypes.QueryAccountRequest{
			Address: address.String(),
		},
	)
	if err != nil {
		slog.Error("Error getting account", "err", err)
		panic(err)
	}
	// register auth interface

	var acc authTypes.AccountI
	err = chainClient.encCfg.InterfaceRegistry.UnpackAny(resp.Account, &acc)
	if err != nil {
		slog.Error("Error unpacking account", "err", err)
		panic(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence()
}
