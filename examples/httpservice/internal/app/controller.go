package app

type Controller struct {
	store map[string]string
}

func NewController() (*Controller, error) {
	return &Controller{
		store: make(map[string]string),
	}, nil
}
