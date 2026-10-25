package sysproxy

import "errors"

var ErrUnsupported = errors.New("sysproxy: automatic system proxy is not supported here")