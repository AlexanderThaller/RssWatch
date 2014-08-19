package logger

import "errors"

func checkPriority(pr Priority) (err error) {
	_, m := priorities[pr]
	if !m {
		err = errors.New("priority does not exist")
		return
	}

	return
}
