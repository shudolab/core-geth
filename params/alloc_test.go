package params

import (
	"fmt"

	"github.com/shudolab/core-geth/common"
	"github.com/shudolab/core-geth/params/types/genesisT"
)

func ExamplemainnetAllocData() {
	// Test that the mainnet alloc is parsable.
	alloc := mainnetAllocData
	ga := genesisT.DecodePreAlloc(alloc)

	fmt.Println(ga[common.Address{0x3}])
	fmt.Println(ga[common.HexToAddress("0x3000000000000000000000000000000000000003")])
	// Output:
	// {[] map[] <nil> 0 []}
	// {[] map[] <nil> 0 []}
}
