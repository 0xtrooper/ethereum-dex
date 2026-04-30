package show

import "dex/service"

type dataOut struct {
	Config *service.Config
	Exists bool
}

func process(svc ConfigLoader) (*dataOut, error) {
	cfg, exists, err := svc.LoadIfExists()
	if err != nil {
		return nil, err
	}
	return &dataOut{Config: cfg, Exists: exists}, nil
}
