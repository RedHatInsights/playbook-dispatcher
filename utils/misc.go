package utils

import (
	"os"
)

func DieOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func PeekSignalChannel(sigs chan os.Signal) os.Signal {
	select {
	case sig := <-sigs:
		return sig
	default:
		return nil
	}
}
