package main

import (
	"DApp/counter"
	"DApp/pkg/config"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 1. 连接到Sepolia测试网络
	// 替换为你的Infura Sepolia API URL
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

	// 2. 准备账户（用于发送交易）
	key := config.GetConfig().Server.PrivateKey
	// 替换为你的私钥（不包含0x前缀）
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("无法将公钥转换为*ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("获取nonce失败: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("获取Gas价格失败: %v", err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("获取链ID失败: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("创建交易签名器失败: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // 不需要发送以太币
	auth.GasLimit = uint64(300000) // 足够的Gas限制
	auth.GasPrice = gasPrice

	// 3. 部署合约（如果尚未部署）
	var contractAddress string
	if contractAddress == "" {
		// 部署新合约
		address, tx, _, err := counter.DeployCounter(auth, client)
		if err != nil {
			log.Fatalf("部署合约失败: %v", err)
		}

		contractAddress = address.Hex()
		fmt.Printf("合约部署中，交易哈希: %s\n", tx.Hash().Hex())
		fmt.Printf("合约地址: %s\n", contractAddress)
		//fmt.Printf("counter: %s\n", instance)

		// 等待部署完成（实际应用中需要轮询确认）
	}

	// 4. 连接到已部署的合约
	contract, err := counter.NewCounter(common.HexToAddress(contractAddress), client)
	if err != nil {
		log.Fatalf("连接到合约失败: %v", err)
	}

	// 5. 调用合约的只读方法（getCount）
	count, err := contract.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("调用getCount失败: %v", err)
	}
	fmt.Printf("当前计数: %d\n", count)

	// 6. 调用合约的写方法（increment）
	tx, err := contract.Increment(auth)
	if err != nil {
		log.Fatalf("调用increment失败: %v", err)
	}
	fmt.Printf("增加计数的交易已发送，哈希: %s\n", tx.Hash().Hex())

	// 等待交易确认后再次查询（实际应用中需要轮询确认）
	// 这里为了演示，简单等待几秒后查询
	fmt.Println("等待交易确认...")
	// time.Sleep(60 * time.Second)

	// 再次查询计数
	newCount, err := contract.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("再次调用getCount失败: %v", err)
	}
	fmt.Printf("增加后的计数: %d\n", newCount)
}
