package chainid

type dataOut struct {
	ChainID    int64
	ConfigPath string
	Created    bool
}

func process(svc ConfigService, in *dataIn) (*dataOut, error) {
	cfg, created, err := svc.Ensure()
	if err != nil {
		return nil, err
	}
	cfg.Network.ChainID = in.ChainID
	if err := svc.Save(cfg); err != nil {
		return nil, err
	}
	return &dataOut{ChainID: cfg.Network.ChainID, ConfigPath: svc.Path(), Created: created}, nil
}
