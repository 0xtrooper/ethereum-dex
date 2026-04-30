package walletexport

type dataOut struct {
	PrivateKey string
}

func process(ks WalletExporter, in *dataIn) (*dataOut, error) {
	privateKey, err := ks.Export(in.Address, in.Password)
	if err != nil {
		return nil, err
	}
	return &dataOut{PrivateKey: privateKey}, nil
}
