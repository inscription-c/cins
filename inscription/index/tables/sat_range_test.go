package tables

import (
	"fmt"
	"testing"
)

func TestSatRange(t *testing.T) {
	satRange := SatRange{
		Start: 0,
		End:   50000000000,
	}
	fmt.Println(satRange.Store())
	fmt.Println(NewSatRange(satRange.Store()))
}
