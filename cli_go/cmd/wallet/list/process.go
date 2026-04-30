package walletlist

type dataOut struct {
	Addresses []string
}

func process(ks WalletLister) (*dataOut, error) {
	return &dataOut{Addresses: ks.List()}, nil
}
