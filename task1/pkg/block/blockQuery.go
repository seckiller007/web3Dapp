package main

import (
	"DApp/pkg/config"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 替换为你的Infura Sepolia API URL  注意InitConfig配置时 工作目录设置一致，否则有的启动类etc/config.yaml这个路径会加载不到。
	config.InitConfig("etc/config.yaml")
	apIkey := config.GetConfig().Server.APIkey
	infuraURL := "https://sepolia.infura.io/v3/" + apIkey

	// 连接到Sepolia测试网络
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("无法连接到以太坊客户端: %v", err)
	}
	defer client.Close()

	fmt.Println("成功连接到Sepolia测试网络")

	// 要查询的区块号，这里使用最新区块
	blockNumber := big.NewInt(-1) // -1表示最新区块

	// 获取区块信息
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatalf("无法获取区块: %v", err)
	}

	// 打印区块信息
	fmt.Printf("区块号: %d\n", block.Number().Uint64())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希: %s\n", block.ParentHash().Hex())
	fmt.Printf("时间戳: %v\n", block.Time())
	fmt.Printf("交易数量: %d\n", len(block.Transactions()))
	fmt.Printf("难度: %s\n", block.Difficulty().String())
	fmt.Printf("Gas上限: %d\n", block.GasLimit())
	fmt.Printf("Gas使用: %d\n", block.GasUsed())
	fmt.Printf("矿工地址: %s\n", block.Coinbase().Hex())
	key := config.GetConfig().Server.PrivateKey
	// 替换为你的私钥（不包含0x前缀）
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Fatalf("无法解析私钥: %v", err)
	}

	// 从私钥获取公钥
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("无法将公钥转换为*ecdsa.PublicKey")
	}

	// 从公钥获取发送者地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取发送者的nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("无法获取nonce: %v", err)
	}

	// 转账金额 (0.01 ETH)
	amount := big.NewInt(10000000000000000) // 1 ETH = 1e18 wei

	// 替换为接收者地址
	toAddress := common.HexToAddress("0x78090ebB7d05CdAFAfD8953b5358D04C26865582")

	// 获取当前Gas价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("无法获取Gas价格: %v", err)
	}

	// 替换原来的gasLimit估计代码
	//callMsg := &types.Message{
	//	From:     fromAddress,
	//	To:       &toAddress,
	//	Value:    amount,
	//	Gas:      0, // 留空让节点估算
	//	GasPrice: gasPrice,
	//	Data:     nil, // 普通转账不需要数据
	//}

	// 使用自定义的CallMsg结构体
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Value: amount,
	})
	if err != nil {
		log.Fatalf("无法估计Gas限制: %v", err)
	}

	// 获取当前链ID (Sepolia的链ID是11155111)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("无法获取链ID: %v", err)
	}

	// 构建交易
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("无法签名交易: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	// 输出交易哈希
	fmt.Printf("交易已发送，哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("可以在Etherscan查看: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())
}
