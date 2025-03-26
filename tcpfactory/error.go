package tcpfactory

import "errors"

var (
	ErrHostNotValid = errors.New("host not valid")
	ErrPortNotValid = errors.New("port not valid")
)
