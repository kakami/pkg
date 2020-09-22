package aqc_test

import (
	"testing"

	"pkg/etcd/aqc"
)

func Test_AcQueue(t *testing.T) {
	aqc.NewAcQueue(nil, nil)
}
