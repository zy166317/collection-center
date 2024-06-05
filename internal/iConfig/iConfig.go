package iConfig

type Rpc struct {
	Test         bool
	EthRpc       *[]string
	BtcRpc       *[]string
	RemoteSigner *RemoteSigner
}

type RemoteSigner struct {
	Host       string
	Port       string
	TlsPemPath string
	User       string
	Pass       string
}
