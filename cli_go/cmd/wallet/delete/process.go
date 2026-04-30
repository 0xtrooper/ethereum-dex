package walletdelete

type result struct {
	Address string
	Err     error
}

type dataOut struct {
	Results []result
}

func process(ks WalletDeleter, in *dataIn) (*dataOut, error) {
	results := make([]result, len(in.Addresses))
	for i, addr := range in.Addresses {
		results[i] = result{
			Address: addr,
			Err:     ks.Delete(addr),
		}
	}
	return &dataOut{Results: results}, nil
}
