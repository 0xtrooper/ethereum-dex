package configterminate

type dataOut struct {
	Path string
}

func process(svc ConfigTerminator, in *dataIn) (*dataOut, error) {
	if err := svc.Delete(); err != nil {
		return nil, err
	}
	return &dataOut{Path: in.Path}, nil
}
