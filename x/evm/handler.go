package evm

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/cosmos/ethermint/zktx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tmtypes "github.com/tendermint/tendermint/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgEthereumTx:
			return handleMsgEthereumTx(ctx, k, msg)
		case types.MsgEthermint:
			return handleMsgEthermint(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// handleMsgEthereumTx handles an Ethereum specific tx
func handleMsgEthereumTx(ctx sdk.Context, k Keeper, msg types.MsgEthereumTx) (*sdk.Result, error) {
	// parse the chainID from a string to a base-10 integer
	fmt.Println("handle",msg.From())
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	// Verify signature and retrieve sender address
	sender, err := msg.VerifySig(chainIDEpoch)
	if err != nil {
		return nil, err
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := common.BytesToHash(txHash)

	st := types.StateTransition{
		AccountNonce: msg.Data.AccountNonce,
		Price:        msg.Data.Price,
		GasLimit:     msg.Data.GasLimit,
		Recipient:    msg.Data.Recipient,
		Amount:       msg.Data.Amount,
		Payload:      msg.Data.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		// TODO: block hash
		k.CommitStateDB.Prepare(ethHash, common.Hash{}, k.TxCount)
		k.TxCount++
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}


	//add for blockmaze just like applyTrsaction
	initSN := zktx.ComputePRF(zktx.ZKTxAddress.Hash().Bytes(), common.Hash{}.Bytes())
	statedb := k.CommitStateDB
	if msg.TxCode() == types.MintTx { //
		if exist := statedb.Exist(common.BytesToAddress(msg.ZKSN().Bytes())); exist == true && (*(msg.ZKSN()) != *(initSN)) { //if sn is already exist,
			return nil,errors.New("sn is already used")
		}
		cmtbalance := k.GetCMTBalance(common.BytesToAddress(msg.From()))
		if err = zktx.VerifyMintProof(&cmtbalance, msg.ZKSN(), msg.ZKCMT(), msg.ZKValue(), msg.ZKProof()); err != nil {
			fmt.Println("invalid zk mint proof: ", err)
			return nil, err
		}
		statedb.CreateAccount(common.BytesToAddress(msg.ZKSN().Bytes()))
		statedb.SetNonce(common.BytesToAddress(msg.ZKSN().Bytes()), 1)
	} else if msg.TxCode() == types.SendTx {
		cmtbalance := k.GetCMTBalance(common.BytesToAddress(msg.From()))
		if exist := statedb.Exist(common.BytesToAddress(msg.ZKSN().Bytes())); exist == true && (*(msg.ZKSN()) != *(initSN)) { //if sn is already exist,
			return nil, errors.New("sn is already used ")
		}
		if err = zktx.VerifySendProof(msg.ZKSN(), msg.ZKCMTS(), msg.ZKProof(), &cmtbalance, msg.ZKCMT()); err != nil {
			fmt.Println("invalid zk send proof: ", err)
			return nil, err
		}
		statedb.CreateAccount(common.BytesToAddress(msg.ZKSN().Bytes()))
		statedb.SetNonce(common.BytesToAddress(msg.ZKSN().Bytes()), 1)
		// } else if tx.TxCode() == types.UpdateTx {
		// 	cmtbalance := statedb.GetCMTBalance(msg.From())
		// 	if err = zktx.VerifyUpdateProof(&cmtbalance, tx.RTcmt(), tx.ZKCMT(), tx.ZKProof()); err != nil {
		// 		fmt.Println("invalid zk update proof: ", err)
		// 		return nil, 0, err
		// 	}
	} else if msg.TxCode() == types.DepositTx {
		if exist := statedb.Exist(common.BytesToAddress(msg.ZKSN().Bytes())); exist == true && (*(msg.ZKSN()) != *(initSN)) { //if sn is already exist,
			return nil, errors.New("sn in deposit tx has been already used")
		}
		cmtbalance := k.GetCMTBalance(common.BytesToAddress(msg.From()))
		addr1, err := types.ExtractPKBAddress(ethtypes.HomesteadSigner{}, &msg) //tbd
		ppp := ecdsa.PublicKey{crypto.S256(), msg.X(), msg.Y()}
		addr2 := crypto.PubkeyToAddress(ppp)
		if err != nil || addr1 != addr2 {
			return nil, errors.New("invalid depositTx signature ")
		}
		if err = zktx.VerifyDepositProof(&ppp, msg.RTcmt(), &cmtbalance, msg.ZKSN(), msg.ZKCMT(), msg.ZKSNS(), msg.ZKProof()); err != nil {
			fmt.Println("invalid zk deposit proof: ", err)
			return nil,  err
		}
		statedb.CreateAccount(common.BytesToAddress(msg.ZKSN().Bytes()))
		statedb.SetNonce(common.BytesToAddress(msg.ZKSN().Bytes()), 1)
	} else if msg.TxCode() == types.RedeemTx {
		if exist := statedb.Exist(common.BytesToAddress(msg.ZKSN().Bytes())); exist == true && (*(msg.ZKSN()) != *(initSN)) { //if sn is already exist,
			return nil, errors.New("sn is already used ")
		}
		cmtbalance := k.GetCMTBalance(common.BytesToAddress(msg.From()))
		if err = zktx.VerifyRedeemProof(&cmtbalance, msg.ZKSN(), msg.ZKCMT(), msg.ZKValue(), msg.ZKProof()); err != nil {
			fmt.Println("invalid zk redeem proof: ", err)
			return nil, err
		}
		statedb.CreateAccount(common.BytesToAddress(msg.ZKSN().Bytes()))
		statedb.SetNonce(common.BytesToAddress(msg.ZKSN().Bytes()), 1)
	}
	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		return nil, err
	}

	if !st.Simulate {
		// update block bloom filter
		k.Bloom.Or(k.Bloom, executionResult.Bloom)

		// update transaction logs in KVStore
		err = k.SetLogs(ctx, common.BytesToHash(txHash), executionResult.Logs)
		if err != nil {
			panic(err)
		}
	}

	// log successful execution
	k.Logger(ctx).Info(executionResult.Result.Log)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthereumTx,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Data.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	if msg.Data.Recipient != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, msg.Data.Recipient.String()),
			),
		)
	}

	// set the events to the result
	executionResult.Result.Events = ctx.EventManager().Events()
	return executionResult.Result, nil
}

// handleMsgEthermint handles an sdk.StdTx for an Ethereum state transition
func handleMsgEthermint(ctx sdk.Context, k Keeper, msg types.MsgEthermint) (*sdk.Result, error) {
	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := common.BytesToHash(txHash)

	st := types.StateTransition{
		AccountNonce: msg.AccountNonce,
		Price:        msg.Price.BigInt(),
		GasLimit:     msg.GasLimit,
		Amount:       msg.Amount.BigInt(),
		Payload:      msg.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       common.BytesToAddress(msg.From.Bytes()),
		Simulate:     ctx.IsCheckTx(),
	}

	if msg.Recipient != nil {
		to := common.BytesToAddress(msg.Recipient.Bytes())
		st.Recipient = &to
	}

	if !st.Simulate {
		// Prepare db for logs
		k.CommitStateDB.Prepare(ethHash, common.Hash{}, k.TxCount)
		k.TxCount++
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		return nil, err
	}

	// update block bloom filter
	if !st.Simulate {
		k.Bloom.Or(k.Bloom, executionResult.Bloom)

		// update transaction logs in KVStore
		err = k.SetLogs(ctx, common.BytesToHash(txHash), executionResult.Logs)
		if err != nil {
			panic(err)
		}
	}

	// log successful execution
	k.Logger(ctx).Info(executionResult.Result.Log)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthermint,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	})

	if msg.Recipient != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthermint,
				sdk.NewAttribute(types.AttributeKeyRecipient, msg.Recipient.String()),
			),
		)
	}

	// set the events to the result
	executionResult.Result.Events = ctx.EventManager().Events()
	return executionResult.Result, nil
}
