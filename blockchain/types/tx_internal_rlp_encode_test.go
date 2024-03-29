// Copyright 2019 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

//+build RLPEncodeTest

package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/klaytn/klaytn/blockchain/types/accountkey"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/crypto"
	"github.com/klaytn/klaytn/params"
	"github.com/klaytn/klaytn/ser/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

var payerKey = defaultFeePayerKey()

// TestTxRLPEncode tests encoding transactions.
func TestTxRLPEncode(t *testing.T) {
	funcs := []testingF{
		testTxRLPEncodeLegacy,

		testTxRLPEncodeValueTransfer,
		testTxRLPEncodeFeeDelegatedValueTransfer,
		testTxRLPEncodeFeeDelegatedValueTransferWithRatio,

		testTxRLPEncodeValueTransferMemo,
		testTxRLPEncodeFeeDelegatedValueTransferMemo,
		testTxRLPEncodeFeeDelegatedValueTransferMemoWithRatio,

		testTxRLPEncodeAccountUpdate,
		testTxRLPEncodeFeeDelegatedAccountUpdate,
		testTxRLPEncodeFeeDelegatedAccountUpdateWithRatio,

		testTxRLPEncodeSmartContractDeploy,
		testTxRLPEncodeFeeDelegatedSmartContractDeploy,
		testTxRLPEncodeFeeDelegatedSmartContractDeployWithRatio,

		testTxRLPEncodeSmartContractExecution,
		testTxRLPEncodeFeeDelegatedSmartContractExecution,
		testTxRLPEncodeFeeDelegatedSmartContractExecutionWithRatio,

		testTxRLPEncodeCancel,
		testTxRLPEncodeFeeDelegatedCancel,
		testTxRLPEncodeFeeDelegatedCancelWithRatio,

		testTxRLPEncodeChainDataAnchoring,
	}

	for _, f := range funcs {
		fnname := getFunctionName(f)
		fnname = fnname[strings.LastIndex(fnname, ".")+1:]
		t.Run(fnname, func(t *testing.T) {
			f(t)
		})
	}
}

func printRLPEncode(chainId *big.Int, signer Signer, sigRLP *bytes.Buffer, txHashRLP *bytes.Buffer, senderTxHashRLP *bytes.Buffer, rawTx *Transaction) {
	privateKey := crypto.FromECDSA(key)

	vrs, _ := rlp.EncodeToBytes(rawTx.data.RawSignatureValues())

	fmt.Printf("ChainID %#x\n", chainId)
	fmt.Printf("PrivateKey %#x\n", privateKey)
	fmt.Printf("PublicKey.X %#x\n", key.X)
	fmt.Printf("PublicKey.Y %#x\n", key.Y)
	fmt.Printf("SigRLP %#x\n", sigRLP.Bytes())
	fmt.Printf("SigHash %s\n", signer.Hash(rawTx).String())
	fmt.Printf("Signature %s\n", common.Bytes2Hex(vrs))
	fmt.Printf("TxHashRLP %#x\n", txHashRLP.Bytes())
	fmt.Printf("TxHash %#x\n", rawTx.Hash())
	fmt.Printf("SenderTxHashRLP %#x\n", senderTxHashRLP.Bytes())
	fmt.Printf("SenderTxHash %#x\n", rawTx.SenderTxHashAll())
	fmt.Println(rawTx)

}

func printFeeDelegatedRLPEncode(t *testing.T, chainId *big.Int, signer Signer, sigRLP *bytes.Buffer, feePayerSigRLP *bytes.Buffer, txHashRLP *bytes.Buffer, senderTxHashRLP *bytes.Buffer, rawTx *Transaction) {
	privateKey := crypto.FromECDSA(key)
	vrs, _ := rlp.EncodeToBytes(rawTx.data.RawSignatureValues())

	fmt.Printf("ChainID %#x\n", chainId)
	// Sender
	fmt.Printf("PrivateKey %#x\n", privateKey)
	fmt.Printf("PublicKey.X %#x\n", key.X)
	fmt.Printf("PublicKey.Y %#x\n", key.Y)
	fmt.Printf("SigRLP %#x\n", sigRLP.Bytes())
	fmt.Printf("SigHash %s\n", signer.Hash(rawTx).String())
	fmt.Printf("Signature %s\n", common.Bytes2Hex(vrs))

	// FeePayer
	feePayerPrivateKey := crypto.FromECDSA(payerKey)

	feePayerHash, err := signer.HashFeePayer(rawTx)
	assert.Equal(t, nil, err)

	feePayerVrs, _ := rlp.EncodeToBytes(rawTx.data.(TxInternalDataFeePayer).GetFeePayerRawSignatureValues())

	fmt.Printf("FeePayerPrivateKey %#x\n", feePayerPrivateKey)
	fmt.Printf("FeePayerPublicKey.X %#x\n", payerKey.X)
	fmt.Printf("FeePayerPublicKey.Y %#x\n", payerKey.Y)
	fmt.Printf("SigRLPFeePayer %#x\n", feePayerSigRLP.Bytes())
	fmt.Printf("SigHashFeePayer %s\n", feePayerHash.String())
	fmt.Printf("SignatureFeePayer %s\n", common.Bytes2Hex(feePayerVrs))

	fmt.Printf("TxHashRLP %#x\n", txHashRLP.Bytes())
	fmt.Printf("TxHash %#x\n", rawTx.Hash())
	fmt.Printf("SenderTxHashRLP %#x\n", senderTxHashRLP.Bytes())
	fmt.Printf("SenderTxHash %#x\n", rawTx.SenderTxHashAll())
	fmt.Println(rawTx)
}

func testTxRLPEncodeLegacy(t *testing.T) {
	tx := genLegacyTransaction().(*TxInternalDataLegacy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)
	err := rlp.Encode(sigRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.Payload,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.Payload,
		tx.V,
		tx.R,
		tx.S,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeValueTransfer(t *testing.T) {
	tx := genValueTransferTransaction().(*TxInternalDataValueTransfer)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeValueTransferMemo(t *testing.T) {
	tx := genValueTransferMemoTransaction().(*TxInternalDataValueTransferMemo)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

//func testTxRLPEncodeAccountCreation(t *testing.T) {
//	tx := genAccountCreationTransaction().(*TxInternalDataAccountCreation)
//
//	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
//	chainId := params.BFTTestChainConfig.ChainID
//	rawTx := &Transaction{data: tx}
//	rawTx.Sign(signer, key)
//
//	sigRLP := new(bytes.Buffer)
//
//	err := rlp.Encode(sigRLP, []interface{}{
//		tx.SerializeForSignToBytes(),
//		chainId,
//		uint(0),
//		uint(0),
//	})
//	assert.Equal(t, nil, err)
//
//	txHashRLP := new(bytes.Buffer)
//	err = rlp.Encode(txHashRLP, tx.Type())
//	assert.Equal(t, nil, err)
//
//	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
//	keyEnc, _ := rlp.EncodeToBytes(serializer)
//
//	err = rlp.Encode(txHashRLP, []interface{}{
//		tx.AccountNonce,
//		tx.Price,
//		tx.GasLimit,
//		tx.Recipient,
//		tx.Amount,
//		tx.From,
//		tx.HumanReadable,
//		keyEnc,
//		tx.TxSignatures,
//	})
//	assert.Equal(t, nil, err)
//
//	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
//}

func testTxRLPEncodeAccountUpdate(t *testing.T) {
	tx := genAccountUpdateTransaction().(*TxInternalDataAccountUpdate)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeSmartContractDeploy(t *testing.T) {
	tx := genSmartContractDeployTransaction().(*TxInternalDataSmartContractDeploy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.CodeFormat,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeSmartContractExecution(t *testing.T) {
	tx := genSmartContractExecutionTransaction().(*TxInternalDataSmartContractExecution)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeCancel(t *testing.T) {
	tx := genCancelTransaction().(*TxInternalDataCancel)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeChainDataAnchoring(t *testing.T) {
	tx := genChainDataTransaction().(*TxInternalDataChainDataAnchoring)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransfer(t *testing.T) {
	tx := genFeeDelegatedValueTransferTransaction().(*TxInternalDataFeeDelegatedValueTransfer)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferMemo(t *testing.T) {
	tx := genFeeDelegatedValueTransferMemoTransaction().(*TxInternalDataFeeDelegatedValueTransferMemo)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedAccountUpdate(t *testing.T) {
	tx := genFeeDelegatedAccountUpdateTransaction().(*TxInternalDataFeeDelegatedAccountUpdate)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractDeploy(t *testing.T) {
	tx := genFeeDelegatedSmartContractDeployTransaction().(*TxInternalDataFeeDelegatedSmartContractDeploy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.CodeFormat,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.CodeFormat,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractExecution(t *testing.T) {
	tx := genFeeDelegatedSmartContractExecutionTransaction().(*TxInternalDataFeeDelegatedSmartContractExecution)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedCancel(t *testing.T) {
	tx := genFeeDelegatedCancelTransaction().(*TxInternalDataFeeDelegatedCancel)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferWithRatio(t *testing.T) {
	tx := genFeeDelegatedValueTransferWithRatioTransaction().(*TxInternalDataFeeDelegatedValueTransferWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferMemoWithRatio(t *testing.T) {
	tx := genFeeDelegatedValueTransferMemoWithRatioTransaction().(*TxInternalDataFeeDelegatedValueTransferMemoWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedAccountUpdateWithRatio(t *testing.T) {
	tx := genFeeDelegatedAccountUpdateWithRatioTransaction().(*TxInternalDataFeeDelegatedAccountUpdateWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.FeeRatio,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractDeployWithRatio(t *testing.T) {
	tx := genFeeDelegatedSmartContractDeployWithRatioTransaction().(*TxInternalDataFeeDelegatedSmartContractDeployWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.FeeRatio,
		tx.CodeFormat,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.FeeRatio,
		tx.CodeFormat,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractExecutionWithRatio(t *testing.T) {
	tx := genFeeDelegatedSmartContractExecutionWithRatioTransaction().(*TxInternalDataFeeDelegatedSmartContractExecutionWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.FeeRatio,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedCancelWithRatio(t *testing.T) {
	tx := genFeeDelegatedCancelWithRatioTransaction().(*TxInternalDataFeeDelegatedCancelWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})
	assert.Equal(t, nil, err)

	senderTxHashRLP := new(bytes.Buffer)
	err = rlp.Encode(senderTxHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(senderTxHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
	})
	assert.Equal(t, nil, err)

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, senderTxHashRLP, rawTx)
}

func defaultFeePayerKey() *ecdsa.PrivateKey {
	key, _ := crypto.HexToECDSA("b9d5558443585bca6f225b935950e3f6e69f9da8a5809a83f51c3365dff53936")
	return key
}
