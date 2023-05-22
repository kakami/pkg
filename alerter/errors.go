package alerter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	_ErrWarning = fmt.Errorf("this is a warning error")
	_ErrFatal   = fmt.Errorf("this is a fatal error")
)

func Fatal(err error) error {
	select {
	case errPool.fch <- err:
	default:
	}
	return errors.Join(err, _ErrFatal)
}

func IsFatalError(err error) bool {
	return errors.Is(err, _ErrFatal)
}

func IsWarningError(err error) bool {
	return errors.Is(err, _ErrWarning)
}

func Warning(err error) error {
	select {
	case errPool.ech <- err:
	default:
	}
	return errors.Join(err, _ErrWarning)
}

const _AlertTemplate = `{
    "msgtype": "markdown",
    "markdown": {
    "content": "## %s
>ID: %s
>Accrued: %d
>ErrMsg: %s
>Time: %v"
    }
}`

func SendAlert(title string, num int, msg string) {
	context := fmt.Sprintf(_AlertTemplate, title, ID, num, msg, time.Now().Format("2006-01-02 15:04:05"))
	resp, err := http.Post(
		Webhook,
		"application/json",
		bytes.NewBuffer([]byte(context)))
	if err != nil || resp.StatusCode != http.StatusOK {
		resp, err = http.Post(
			Webhook,
			"application/json",
			bytes.NewBuffer([]byte(context)))
	}
	if err == nil {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}
