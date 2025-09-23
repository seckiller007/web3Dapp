package config

const (
	DevNetRPC  = "https://api.devnet.solana.com"
	DevNetWS   = "wss://api.devnet.solana.com"
	MainNetRPC = "https://api.mainnet-beta.solana.com"
	MainNetWS  = "wss://api.mainnet-beta.solana.com"
	TestNetRPC = "https://api.testnet.solana.com"
	TestNetWS  = "wss://api.testnet.solana.com"
)

type Config struct {
	Network     string
	RPCEndpoint string
	WSEndpoint  string
	PrivateKey  string
}

func GetConfig(network string) Config {
	switch network {
	case "mainnet":
		return Config{
			Network:     "mainnet",
			RPCEndpoint: MainNetRPC,
			WSEndpoint:  MainNetWS,
		}
	case "testnet":
		return Config{
			Network:     "testnet",
			RPCEndpoint: TestNetRPC,
			WSEndpoint:  TestNetWS,
		}
	default: // devnet
		return Config{
			Network:     "devnet",
			RPCEndpoint: DevNetRPC,
			WSEndpoint:  DevNetWS,
		}
	}
}
