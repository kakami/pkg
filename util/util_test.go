package util_test

import (
	"testing"

	"github.com/kakami/pkg/util"
)

func Test_Prime(t *testing.T) {
	t.Log(util.Prime(1))
	t.Log(util.Prime(2))
	t.Log(util.Prime(3))
	t.Log(util.Prime(4))
	t.Log(util.Prime(12))
	t.Log(util.Prime(123))
	t.Log(util.Prime(1234))
}
