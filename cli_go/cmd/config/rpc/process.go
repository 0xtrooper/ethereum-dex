package rpc

type dataOut struct {
	RPCURL     string
	ConfigPath string
	Created    bool
}

func process(svc ConfigService, in *dataIn) (*dataOut, error) {
	cfg, created, err := svc.Ensure()
	if err != nil {
		return nil, err
	}
	cfg.Network.RPCURL = in.RPCURL
	if err := svc.Save(cfg); err != nil {
		return nil, err
	}
	return &dataOut{RPCURL: cfg.Network.RPCURL, ConfigPath: svc.Path(), Created: created}, nil
}
