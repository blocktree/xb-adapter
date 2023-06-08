/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwtester

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
	"path/filepath"
	"testing"
)

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
}

// BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Notice("header:", header)
	return nil
}

// BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	//for i, input := range data.TxInputs {
	//	log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	//}
	//
	//for i, output := range data.TxOutputs {
	//	log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	//}
	//
	//log.Std.Notice("data.Transaction: %+v", data.Transaction)

	return nil
}

// BlockExtractSmartContractDataNotify 区块提取智能合约交易结果通知
func (sub *subscriberSingle) BlockExtractSmartContractDataNotify(sourceKey string, data *openwallet.SmartContractReceipt) error {

	log.Notice("sourceKey:", sourceKey)
	//log.Std.Notice("data.ContractTransaction: %+v", data)

	for i, event := range data.Events {
		//log.Std.Notice("data.Events[%d]: %+v", i, event)
		assetsMgr, err := openw.GetAssetsAdapter(data.Coin.Symbol)
		if err != nil {
			log.Error(data.Coin.Symbol, "is not support")
			return nil
		}
		nftTx, _ := assetsMgr.GetNFTContractDecoder().GetNFTTransfer(event)
		if nftTx != nil {
			log.Std.Notice("NFT Transfer[%d]: %+v", i, nftTx)
		}
	}

	return nil
}

func TestSubscribeAddress_XB(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "XB"
	)

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}
	scanner.SetBlockScanTargetFuncV2(testScanTargetFunc(symbol))
	scanner.Run()

	<-endRunning
}

func TestBlockScanner_ExtractTransactionAndReceiptData(t *testing.T) {

	var (
		symbol = "XB"
		txid   = "0x770324b9ea445e469c874af0d95afa41f217816681c4a59d2d9b8cd503394285"
	)

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}
	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	result, contractResult, err := scanner.ExtractTransactionAndReceiptData(txid, testScanTargetFunc(symbol))
	if err != nil {
		t.Errorf("ExtractTransactionData unexpected error %v", err)
		return
	}

	for sourceKey, keyData := range result {
		log.Notice("account:", sourceKey)
		for _, data := range keyData {

			for i, input := range data.TxInputs {
				log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
			}

			for i, output := range data.TxOutputs {
				log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
			}

			log.Std.Notice("data.Transaction: %+v", data.Transaction)
		}
	}

	for sourceKey, keyData := range contractResult {
		log.Notice("sourceKey:", sourceKey)
		//log.Std.Notice("data.ContractTransaction: %+v", keyData)

		for i, event := range keyData.Events {
			//log.Std.Notice("data.Contract[%d]: %+v", i, event.Contract)
			//log.Std.Notice("data.Events[%d]: %+v", i, event)

			nftTx, _ := assetsMgr.GetNFTContractDecoder().GetNFTTransfer(event)
			if nftTx != nil {
				log.Std.Notice("NFT Transfer[%d]: %+v", i, nftTx)
				for _, token := range nftTx.Tokens {
					metaData, err := assetsMgr.GetNFTContractDecoder().GetMetaDataOfNFT(&token)
					if err != nil {
						continue
					}
					log.Std.Notice("MetaData: %s", metaData.URI)
				}

			}
		}
	}
}

func testScanTargetFunc(symbol string) openwallet.BlockScanTargetFuncV2 {
	var (
		addrs     = make(map[string]openwallet.ScanTargetResult)
		contracts = make(map[string]openwallet.ScanTargetResult)
	)

	//添加监听的合约地址
	contract := &openwallet.SmartContract{
		Symbol:   symbol,
		Address:  "0x2953399124F0cBB46d2CbACD8A89cF0599974963",
		Decimals: 0,
	}
	contract.ContractID = openwallet.GenContractID(contract.Symbol, contract.Address)
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":false,"internalType":"bool","name":"approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256[]","name":"ids","type":"uint256[]"},{"indexed":false,"internalType":"uint256[]","name":"values","type":"uint256[]"}],"name":"TransferBatch","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"TransferSingle","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"value","type":"string"},{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"}],"name":"URI","type":"event"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"id","type":"uint256"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint256[]","name":"ids","type":"uint256[]"}],"name":"balanceOfBatch","outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"address","name":"operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256[]","name":"ids","type":"uint256[]"},{"internalType":"uint256[]","name":"amounts","type":"uint256[]"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"safeBatchTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"bool","name":"approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes4","name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"name":"uri","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`)
	contracts[contract.Address] = openwallet.ScanTargetResult{SourceKey: contract.ContractID, Exist: true, TargetInfo: contract}

	//添加监听的外部地址
	addrs["0x0ff8d979e33412a2904a66226ff6a336d7c873db"] = openwallet.ScanTargetResult{SourceKey: "sender", Exist: true}
	addrs["0xc97a4ed29f03fd549c4ae79086673523122d2bc5"] = openwallet.ScanTargetResult{SourceKey: "receiver", Exist: true}

	scanTargetFunc := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
		if target.ScanTargetType == openwallet.ScanTargetTypeContractAddress {
			return contracts[target.ScanTarget]
		} else if target.ScanTargetType == openwallet.ScanTargetTypeAccountAddress {
			return addrs[target.ScanTarget]
		}
		return openwallet.ScanTargetResult{SourceKey: "", Exist: false, TargetInfo: nil}
	}

	return scanTargetFunc
}

func testBlockScanner(symbol string) openwallet.BlockScanner {
	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return nil
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	if scanner.SupportBlockchainDAI() {
		dbFilePath := filepath.Join("data", "db")
		dbFileName := "blockchain.db"
		file.MkdirAll(dbFilePath)
		dai, err := openwallet.NewBlockchainLocal(filepath.Join(dbFilePath, dbFileName), false)
		if err != nil {
			log.Error("NewBlockchainLocal err: %v", err)
			return nil
		}

		scanner.SetBlockchainDAI(dai)
	}
	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	return scanner
}

func TestScanHeight_XB(t *testing.T) {

	var (
		//endRunning = make(chan bool, 1)
		symbol = "XB"
	)

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}
	scanner.SetBlockScanTargetFuncV2(testScanTargetFunc(symbol))
	scanner.ScanBlock(34888568)
	//scanner.Run()

	//<-endRunning
}
