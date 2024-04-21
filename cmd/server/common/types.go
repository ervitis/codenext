package common

import "context"

type Listener interface {
	Shutdown(context.Context) error
	ListenAndServe() error
}
