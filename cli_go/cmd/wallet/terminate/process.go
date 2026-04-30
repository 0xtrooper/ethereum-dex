package walletterminate

type dataOut struct {
	Count int
}

func process(ks WalletTerminator, in *dataIn) (*dataOut, error) {
	if err := ks.DeleteAll(); err != nil {
		return nil, err
	}
	return &dataOut{Count: in.Count}, nil
}
