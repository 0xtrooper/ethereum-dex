package contact

type dataOut struct {
	Address    string
	ConfigPath string
	Created    bool
}

func process(svc ConfigService, in *dataIn) (*dataOut, error) {
	cfg, created, err := svc.Ensure()
	if err != nil {
		return nil, err
	}
	cfg.Contract.Address = in.Address
	if err := svc.Save(cfg); err != nil {
		return nil, err
	}
	return &dataOut{Address: cfg.Contract.Address, ConfigPath: svc.Path(), Created: created}, nil
}
