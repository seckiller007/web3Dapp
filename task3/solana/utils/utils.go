package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type KeyPairJSON struct {
	SecretKey []byte `json:"secret_key"`
}

// 获取最近区块哈希和最后有效区块高度[8](@ref)
func GetRecentBlockhash(rpcClient *rpc.Client) (solana.Hash, uint64, error) {
	resp, err := rpcClient.GetLatestBlockhash(context.TODO(), rpc.CommitmentConfirmed)
	if err != nil {
		return solana.Hash{}, 0, fmt.Errorf("获取区块哈希失败: %v", err)
	}

	// 计算最后有效区块高度[8](@ref)
	lastValidBlockHeight, err := rpcClient.GetBlockHeight(context.TODO(), rpc.CommitmentConfirmed)
	if err != nil {
		return solana.Hash{}, 0, fmt.Errorf("获取区块高度失败: %v", err)
	}

	return resp.Value.Blockhash, lastValidBlockHeight + 150, nil // 通常有效期为当前高度+150
}

// 查询账户余额
func GetBalance(rpcClient *rpc.Client, account solana.PublicKey) (*rpc.GetBalanceResult, error) {
	balance, err := rpcClient.GetBalance(
		context.TODO(),
		account,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return nil, fmt.Errorf("获取账户余额失败: %w", err)
	}
	return balance, nil
}

// 从私钥获取账户
func GetAccountFromPrivateKey(filePath string) solana.PrivateKey {
	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(filePath)
	if err != nil {
		panic(err)
	}
	//data, err := ioutil.ReadFile("wallet-keypair.json")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//var keypair []byte
	//if err := json.Unmarshal(data, &keypair); err != nil {
	//	log.Fatal(err)
	//}
	//privateKeyBase58 := base58.Encode(keypair[:32])
	return privateKey
}

// 转账SOL[6](@ref)
func TransferSOL(
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	from solana.PrivateKey,
	to solana.PublicKey,
	amount uint64,
	recentBlockhash solana.Hash,
) (solana.Signature, error) {
	// 创建转账指令
	instruction := system.NewTransferInstruction(
		amount,
		from.PublicKey(),
		to,
	).Build()

	// 构造交易
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("构造交易失败: %v", err)
	}

	// 签名交易
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if from.PublicKey().Equals(key) {
			return &from
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("签名交易失败: %v", err)
	}

	// 发送交易并等待确认（带重试机制）
	maxRetries := 3
	var sig solana.Signature
	for i := 0; i < maxRetries; i++ {
		sig, err = confirm.SendAndConfirmTransaction(
			context.TODO(),
			rpcClient,
			wsClient,
			tx,
		)
		if err != nil {
			if shouldRetry(err) {
				log.Printf("Attempt %d/%d failed: %v. Retrying...", i+1, maxRetries, err)
				time.Sleep(2 * time.Second) // 等待后重试
				continue
			}
			log.Fatalf("Failed to send transaction: %v", err)
		}
		break
	}

	return sig, nil
}

// shouldRetry 判断错误是否应该重试
func shouldRetry(err error) bool {
	errMsg := err.Error()
	retryableErrors := []string{
		"Blockhash not found",
		"block height exceeded",
		"BlockhashNotFound",
	}
	for _, retryable := range retryableErrors {
		if strings.Contains(errMsg, retryable) {
			return true
		}
	}
	return false
}

// 监听交易确认[7](@ref)
func WaitForConfirmation(wsClient *ws.Client, signature solana.Signature, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 创建 WebSocket 客户端用于监听交易状态
	wsClient, err := ws.Connect(ctx, rpc.MainNetBeta_WS)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %v", err)
	}
	defer wsClient.Close()

	// 订阅交易签名
	sub, err := wsClient.SignatureSubscribe(
		signature,
		"",
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to signature: %v", err)
	}
	defer sub.Unsubscribe()

	// 等待交易确认
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for confirmation")

		case resp, ok := <-sub.Response():
			if !ok {
				return fmt.Errorf("subscription closed")
			}

			if resp.Value.Err != nil {
				return fmt.Errorf("transaction failed: %v", resp.Value.Err)
			}

			//// 检查交易状态
			//if resp.Value.Context.Slot== rpc.ConfirmationStatusFinalized {
			//	return nil
			//}
		}
	}
}
