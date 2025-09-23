package main

import (
	"context"
	"fmt"
	"log"
	"solana-go/utils"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	// 初始化RPC客户端
	rpcClient := rpc.New(rpc.DevNet_RPC)
	// 4. 监听交易事件
	wsClient, err := ws.Connect(context.Background(), rpc.DevNet_WS)
	if err != nil {
		log.Fatalf("WebSocket连接失败: %v", err)
	}
	defer wsClient.Close()
	// 1. 获取最新区块
	recentBlockhash, lastValidBlockHeight, err := utils.GetRecentBlockhash(rpcClient)
	if err != nil {
		log.Fatalf("获取最新区块失败: %v", err)
	}
	fmt.Printf("最新区块哈希: %s\n", recentBlockhash)
	fmt.Printf("最后有效区块高度: %d\n", lastValidBlockHeight)
	// 3. 构造并发送转账交易

	fromWallet := utils.GetAccountFromPrivateKey("wallet-keypair.json")
	// 替换为实际私钥
	toWallet2 := utils.GetAccountFromPrivateKey("wallet-keypair2.json")

	fmt.Printf("Account 2: %s\n", toWallet2.PublicKey().String())
	// 2. 查询账户余额
	walletAddress := solana.MustPublicKeyFromBase58(fromWallet.PublicKey().String()) // 替换为实际地址
	balance, err := utils.GetBalance(rpcClient, walletAddress)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}
	fmt.Printf("账户余额: %f SOL\n", balance)

	// 替换为实际私钥
	//toWallet := solana.MustPublicKeyFromBase58("Dukdx2R3wMvnAvriU5HnLDFjqgjHDKCqLnux2opC85PT") // 替换为实际地址
	// 替换为实际地址

	amount := uint64(1000000) // 0.001 SOL
	// 请求空投（确保账户有足够余额）
	airdropSig, err := rpcClient.RequestAirdrop(
		context.TODO(),
		fromWallet.PublicKey(),
		solana.LAMPORTS_PER_SOL*1, // 1 SOL
		rpc.CommitmentConfirmed,   // 使用 confirmed 承诺级别
	)
	if err != nil {
		log.Fatalf("Airdrop failed: %v", err)
	}
	fmt.Printf("Airdrop transaction signature: %s\n", airdropSig)
	// 等待空投确认（可选，但建议等待）
	time.Sleep(5 * time.Second)
	//Solana 网络上的每笔交易都必须包含一个最近的区块哈希，它就像一个时间戳，用来确保交易的新鲜度。
	//Solana 的区块哈希有效期很短，通常只有 60-90 秒。如果你的交易在获取区块哈希后没有及时发送并被打包，区块哈希就会因过期而被移出验证节点的队列，从而导致此错误。
	//RPC 节点状态不同步：如果你从一个 RPC 节点获取了最新的区块哈希，但将交易发送到另一个 RPC 节点，而该节点尚未同步到最新的区块状态，它就会无法识别你交易中的区块哈希，从而报错
	// 获取最新区块哈希和有效高度
	recent, err := rpcClient.GetLatestBlockhash(
		context.TODO(),
		rpc.CommitmentConfirmed, // 使用 confirmed 承诺级别
	)
	if err != nil {
		log.Fatalf("Failed to get recent blockhash: %v", err)
	}
	blockhash := recent.Value.Blockhash
	lastlockHeight := recent.Value.LastValidBlockHeight

	// 检查当前区块高度是否超过有效高度
	currentHeight, err := rpcClient.GetBlockHeight(context.TODO(), rpc.CommitmentConfirmed)
	if err != nil {
		log.Fatalf("Failed to get current block height: %v", err)
	}
	if currentHeight > lastlockHeight {
		log.Fatalf("Blockhash expired. Current height: %d, Last valid height: %d", currentHeight, lastlockHeight)
	}
	signature, err := utils.TransferSOL(rpcClient, wsClient, fromWallet, toWallet2.PublicKey(), amount, blockhash)
	if err != nil {
		log.Fatalf("转账失败: %v", err)
	}
	fmt.Printf("转账成功! 交易签名: %s\n", signature)

	sub, err := wsClient.SignatureSubscribe(
		signature,
		"",
	)
	if err != nil {
		log.Fatalf("订阅交易事件失败: %v", err)
	}
	defer sub.Unsubscribe()

	go func() {
		for {
			select {
			case <-sub.Response():
				fmt.Println("交易已确认!")
				return
			case <-time.After(30 * time.Second):
				fmt.Println("交易确认超时")
				return
			}
		}
	}()

	// 5. 智能合约交互示例
	//swapClient, err := tokenswap.NewTokenSwapClient(rpcClient, fromWallet)
	//if err != nil {
	//	log.Fatalf("创建代币交换客户端失败: %v", err)
	//}
	//
	//// 执行代币交换
	//swapSignature, err := swapClient.SwapTokens(
	//	context.Background(),
	//	solana.MustPublicKeyFromBase58("InputTokenMint"),
	//	solana.MustPublicKeyFromBase58("OutputTokenMint"),
	//	uint64(1000000), // 输入金额
	//	recentBlockhash,
	//)
	//if err != nil {
	//	log.Fatalf("代币交换失败: %v", err)
	//}
	//fmt.Printf("代币交换成功! 交易签名: %s\n", swapSignature)
	//
	//// 等待用户输入以保持程序运行
	//fmt.Println("按Enter键退出...")
	//fmt.Scanln()
}
