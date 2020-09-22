package types_test

import (
	"testing"

	"pkg/types"
)

func Test_RoundRobinMembership(t *testing.T) {
	mm := types.NewRoundRobinMembership()
	mm.Add("1")
	mm.Add("2")
	str := "111"
	t.Log(mm.Get(&str))
	t.Log(mm.Get(&str))
	mm.Adds("1", "2", "3", "4")
	t.Log(mm.Members())
	t.Log(mm.Get(&str))
}
