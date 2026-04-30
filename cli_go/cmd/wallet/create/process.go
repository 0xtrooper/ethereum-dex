package walletcreate

type dataOut struct {
	Address string
}

func process(ks WalletCreator, in *dataIn) (*dataOut, error) {
	var (
		address string
		err     error
	)
	if in.Mnemonic != "" {
		address, err = ks.CreateFromMnemonic(in.Mnemonic, in.Password)
	} else {
		address, err = ks.Create(in.Password)
	}
	if err != nil {
		return nil, err
	}
	return &dataOut{Address: address}, nil
}
