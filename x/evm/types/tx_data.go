package types

import (
	"fmt"
	"math/big"

	"github.com/cosmos/ethermint/utils"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// TxData implements the Ethereum transaction data structure. It is used
// solely as intended in Ethereum abiding by the protocol.
// Type of transaction using Code --Agzs 09.17
const (
	PublicTx  uint8 = 0x00
	MintTx    uint8 = 0x01
	SendTx    uint8 = 0x02
	DepositTx uint8 = 0x03
	UpdateTx  uint8 = 0x04
	RedeemTx  uint8 = 0x05
)

type TxData struct {
	AccountNonce uint64          `json:"nonce"`
	Price        *big.Int        `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	Recipient    *ethcmn.Address `json:"to" rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"`
	Payload      []byte          `json:"input"`

	// signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`

	// hash is only used when marshaling to JSON
	Hash *ethcmn.Hash `json:"hash" rlp:"-"`

	//add for blockmaze
	Code uint8 `json:"Code"`

	ZKValue   uint64          `json:"zkvalue"`
	ZKSN      *ethcmn.Hash    `json:"zksn"`
	ZKSNS     *ethcmn.Hash    `json:"zksns"`
	ZKNounce  uint64          `json:"zknounce"`
	ZKAdrress *ethcmn.Address `json:"zkaddress"`
	ZKCMT     *ethcmn.Hash    `json:"zkcmt"`
	ZKCMTS    *ethcmn.Hash    `json:"zkcmts"` //add by zy
	ZKProof   []byte          `json:"zkproof"`
	//	CMTProof  []byte
	RTcmt    ethcmn.Hash `json:"rtcmt"`
	CMTBlock []uint64    `json:"cmtblock"`
	AUX      []byte      `json:"aux"`
	X        *big.Int    `json:"x"`
	Y        *big.Int    `json:"y"`

	//depostiTx signature values
	DepositTxV *big.Int `json:"depositxv"`
	DepositTxR *big.Int `json:"depositxr"`
	DepositTxS *big.Int `json:"depositxs"`
}

// encodableTxData implements the Ethereum transaction data structure. It is used
// solely as intended in Ethereum abiding by the protocol.
type encodableTxData struct {
	AccountNonce uint64          `json:"nonce"`
	Price        string          `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	Recipient    *ethcmn.Address `json:"to" rlp:"nil"` // nil means contract creation
	Amount       string          `json:"value"`
	Payload      []byte          `json:"input"`

	// signature values
	V string `json:"v"`
	R string `json:"r"`
	S string `json:"s"`

	// hash is only used when marshaling to JSON
	Hash *ethcmn.Hash `json:"hash" rlp:"-"`

	//add for blockmaze
	Code uint8 `json:"Code"`

	ZKValue   uint64          `json:"zkvalue"`
	ZKSN      *ethcmn.Hash    `json:"zksn"`
	ZKSNS     *ethcmn.Hash    `json:"zksns"`
	ZKNounce  uint64          `json:"zknounce"`
	ZKAdrress *ethcmn.Address `json:"zkaddress"`
	ZKCMT     *ethcmn.Hash    `json:"zkcmt"`
	ZKCMTS    *ethcmn.Hash    `json:"zmcmts"` //add by zy
	ZKProof   []byte          `json:"zkproof"`
	//	CMTProof  []byte
	RTcmt    ethcmn.Hash `json:"rtcmt"`
	CMTBlock []uint64    `json:"cmtblock"`
	AUX      []byte      `json:"aux"`
	X        string      `json:"x"`
	Y        string      `json:"y"`

	//depostiTx signature values
	DepositTxV string `json:"depositxv"`
	DepositTxR string `json:"depositxr"`
	DepositTxS string `json:"depositxs"`
}

func (td TxData) String() string {
	if td.Recipient != nil || td.ZKSN!=nil {
		return fmt.Sprintf("nonce=%d price=%s gasLimit=%d recipient=%s amount=%s data=0x%x v=%s r=%s s=%s zksn=%s Code=%d zkproof=0x%x" ,
			td.AccountNonce, td.Price, td.GasLimit, td.Recipient.Hex(), td.Amount, td.Payload, td.V, td.R, td.S,td.ZKSN.Hex(),td.Code,td.ZKProof,
			)
	}
	if td.Recipient != nil || td.ZKSN==nil {
		return fmt.Sprintf("nonce=%d price=%s gasLimit=%d recipient=%s amount=%s data=0x%x v=%s r=%s s=%s zksn=nil Code=%d zkproof=0x%x" ,
			td.AccountNonce, td.Price, td.GasLimit, td.Recipient.Hex(), td.Amount, td.Payload, td.V, td.R, td.S,td.Code,td.ZKProof,
		)
	}
	if td.Recipient == nil || td.ZKSN!=nil {
		return fmt.Sprintf("nonce=%d price=%s gasLimit=%d recipient=nil amount=%s data=0x%x v=%s r=%s s=%s zksn=%s Code=%d zkproof=0x%x" ,
			td.AccountNonce, td.Price, td.GasLimit, td.Recipient.Hex(), td.Amount, td.Payload, td.V, td.R, td.S,td.ZKSN.Hex(),td.Code,td.ZKProof,
		)
	}
	return fmt.Sprintf("nonce=%d price=%s gasLimit=%d recipient=nil amount=%s data=0x%x v=%s r=%s s=%s zksn=nil Code=%d zkproof=0x%x",
		td.AccountNonce, td.Price, td.GasLimit, td.Amount, td.Payload, td.V, td.R, td.S,td.Code, td.ZKProof )
}

// MarshalAmino defines custom encoding scheme for TxData
func (td TxData) MarshalAmino() ([]byte, error) {
	gasPrice, err := utils.MarshalBigInt(td.Price)
	if err != nil {
		return nil, err
	}

	amount, err := utils.MarshalBigInt(td.Amount)
	if err != nil {
		return nil, err
	}

	v, err := utils.MarshalBigInt(td.V)
	if err != nil {
		return nil, err
	}

	r, err := utils.MarshalBigInt(td.R)
	if err != nil {
		return nil, err
	}

	s, err := utils.MarshalBigInt(td.S)
	if err != nil {
		return nil, err
	}
	x, err := utils.MarshalBigInt(td.X)
	if err != nil {
		return nil, err
	}
	y, err := utils.MarshalBigInt(td.Y)
	if err != nil {
		return nil, err
	}

	depositTxV, err := utils.MarshalBigInt(td.DepositTxV)
	if err != nil {
		return nil, err
	}
	depositTxR, err := utils.MarshalBigInt(td.DepositTxR)
	if err != nil {
		return nil, err
	}
	depositTxS, err := utils.MarshalBigInt(td.DepositTxS)
	if err != nil {
		return nil, err
	}

	e := encodableTxData{
		AccountNonce: td.AccountNonce,
		Price:        gasPrice,
		GasLimit:     td.GasLimit,
		Recipient:    td.Recipient,
		Amount:       amount,
		Payload:      td.Payload,
		V:            v,
		R:            r,
		S:            s,
		Hash:         td.Hash,

		//blockmaze
		Code:      td.Code,
		ZKValue:   td.ZKValue,
		ZKSN:      td.ZKSN,
		ZKSNS:     td.ZKSNS,
		ZKNounce:  td.ZKNounce,
		ZKAdrress: td.ZKAdrress,
		ZKCMT:     td.ZKCMT,
		ZKCMTS:    td.ZKCMTS, //add by zy
		ZKProof:   td.ZKProof,
		//	CMTProof  []byte
		RTcmt:    td.RTcmt,
		CMTBlock: td.CMTBlock,
		AUX:      td.AUX,
		X:        x,
		Y:        y,

		//depostiTx signature values
		DepositTxV: depositTxV,
		DepositTxR: depositTxR,
		DepositTxS: depositTxS,
	}
	return ModuleCdc.MarshalBinaryBare(e)
}

// UnmarshalAmino defines custom decoding scheme for TxData
func (td *TxData) UnmarshalAmino(data []byte) error {
	var e encodableTxData
	err := ModuleCdc.UnmarshalBinaryBare(data, &e)
	if err != nil {
		return err
	}

	td.AccountNonce = e.AccountNonce
	td.GasLimit = e.GasLimit
	td.Recipient = e.Recipient
	td.Payload = e.Payload
	td.Hash = e.Hash

	price, err := utils.UnmarshalBigInt(e.Price)
	if err != nil {
		return err
	}

	if td.Price != nil {
		td.Price.Set(price)
	} else {
		td.Price = price
	}

	amt, err := utils.UnmarshalBigInt(e.Amount)
	if err != nil {
		return err
	}

	if td.Amount != nil {
		td.Amount.Set(amt)
	} else {
		td.Amount = amt
	}

	v, err := utils.UnmarshalBigInt(e.V)
	if err != nil {
		return err
	}

	if td.V != nil {
		td.V.Set(v)
	} else {
		td.V = v
	}

	r, err := utils.UnmarshalBigInt(e.R)
	if err != nil {
		return err
	}

	if td.R != nil {
		td.R.Set(r)
	} else {
		td.R = r
	}

	s, err := utils.UnmarshalBigInt(e.S)
	if err != nil {
		return err
	}

	if td.S != nil {
		td.S.Set(s)
	} else {
		td.S = s
	}

	//blockmaze


	td.Code = e.Code
	td.ZKValue = e.ZKValue
	td.ZKSN = e.ZKSN
	td.ZKSNS = e.ZKSNS
	td.ZKNounce = e.ZKNounce
	td.ZKAdrress = e.ZKAdrress
	td.ZKCMT = e.ZKCMT

	td.ZKCMTS = e.ZKCMTS //add by zy

	td.ZKProof = e.ZKProof
	//	CMTProof  []byte

	td.RTcmt = e.RTcmt

	td.CMTBlock = e.CMTBlock

	td.AUX = e.AUX

	x, err := utils.UnmarshalBigInt(e.X)
	if err != nil {
		return err
	}
	if td.X != nil {
		td.X.Set(x)
	} else {
		td.X = x
	}

	y, err := utils.UnmarshalBigInt(e.Y)
	if err != nil {
		return err
	}

	if td.Y != nil {
		td.Y.Set(y)
	} else {
		td.Y = y
	}

	depositTxV, err := utils.UnmarshalBigInt(e.DepositTxV)

	if err != nil {
		return err
	}

	if td.DepositTxV != nil {
		td.DepositTxV.Set(depositTxV)
	} else {
		td.DepositTxV = depositTxV
	}

	depositTxR, err := utils.UnmarshalBigInt(e.DepositTxR)
	if err != nil {
		return err
	}

	if td.DepositTxR != nil {
		td.DepositTxR.Set(depositTxR)
	} else {
		td.DepositTxR = depositTxR
	}

	depositTxS, err := utils.UnmarshalBigInt(e.DepositTxS)
	if err != nil {
		return err
	}

	if td.DepositTxS != nil {
		td.DepositTxS.Set(depositTxS)
	} else {
		td.DepositTxS = depositTxS
	}
	fmt.Println(td)
	return nil
}

// TODO: Implement JSON marshaling/ unmarshaling for this type

// TODO: Implement YAML marshaling/ unmarshaling for this type
