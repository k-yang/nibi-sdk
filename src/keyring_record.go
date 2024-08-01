package nibisdk

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (client *ChainClient) GetOrAddAccount(uid string, mnemonic string) *keyring.Record {
	accInfo, err := client.keyring.Key(uid)
	if err == nil {
		return accInfo
	}
	if !sdkerrors.ErrKeyNotFound.Is(err) {
		panic(err)
	}

	accInfo, err = client.keyring.NewAccount(
		uid,
		mnemonic,
		"",
		sdk.FullFundraiserPath,
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}
	return accInfo
}
