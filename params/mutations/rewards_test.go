package mutations

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/shudolab/core-geth/common"
	"github.com/shudolab/core-geth/core/rawdb"
	"github.com/shudolab/core-geth/core/state"
	"github.com/shudolab/core-geth/core/types"
	"github.com/shudolab/core-geth/params"
	"github.com/shudolab/core-geth/params/types/coregeth"
	"github.com/shudolab/core-geth/params/types/ctypes"
	"github.com/holiman/uint256"
)

var (
	defaultEraLength   *big.Int = big.NewInt(5000000)
	MaximumBlockReward          = uint256.NewInt(5e+18)
	WinnerCoinbase              = common.HexToAddress("0000000000000000000000000000000000000001")
	Uncle1Coinbase              = common.HexToAddress("0000000000000000000000000000000000000002")
	Uncle2Coinbase              = common.HexToAddress("0000000000000000000000000000000000000003")

	Era1WinnerReward      = uint256.NewInt(5e+18)               // base block reward
	Era1WinnerUncleReward = uint256.NewInt(156250000000000000)  // uncle inclusion reward (base block reward / 32)
	Era1UncleReward       = uint256.NewInt(4375000000000000000) // uncle reward (depth 1) (block reward * (7/8))

	Era2WinnerReward      = uint256.NewInt(4e+18)
	Era2WinnerUncleReward = new(uint256.Int).Div(uint256.NewInt(4e+18), big32)
	Era2UncleReward       = new(uint256.Int).Div(uint256.NewInt(4e+18), big32)

	Era3WinnerReward      = new(uint256.Int).Mul(new(uint256.Int).Div(Era2WinnerReward, uint256.NewInt(5)), uint256.NewInt(4))
	Era3WinnerUncleReward = new(uint256.Int).Div(new(uint256.Int).Mul(new(uint256.Int).Div(Era2WinnerReward, uint256.NewInt(5)), uint256.NewInt(4)), big32)
	Era3UncleReward       = new(uint256.Int).Div(new(uint256.Int).Mul(new(uint256.Int).Div(Era2WinnerReward, uint256.NewInt(5)), uint256.NewInt(4)), big32)

	Era4WinnerReward      = new(uint256.Int).Mul(new(uint256.Int).Div(Era3WinnerReward, uint256.NewInt(5)), uint256.NewInt(4))
	Era4WinnerUncleReward = new(uint256.Int).Div(new(uint256.Int).Mul(new(uint256.Int).Div(Era3WinnerReward, uint256.NewInt(5)), uint256.NewInt(4)), big32)
	Era4UncleReward       = new(uint256.Int).Div(new(uint256.Int).Mul(new(uint256.Int).Div(Era3WinnerReward, uint256.NewInt(5)), uint256.NewInt(4)), big32)
)

func TestGetBlockEra1(t *testing.T) {
	cases := map[*big.Int]*big.Int{
		big.NewInt(0):         big.NewInt(0),
		big.NewInt(1):         big.NewInt(0),
		big.NewInt(1914999):   big.NewInt(0),
		big.NewInt(1915000):   big.NewInt(0),
		big.NewInt(1915001):   big.NewInt(0),
		big.NewInt(4999999):   big.NewInt(0),
		big.NewInt(5000000):   big.NewInt(0),
		big.NewInt(5000001):   big.NewInt(1),
		big.NewInt(9999999):   big.NewInt(1),
		big.NewInt(10000000):  big.NewInt(1),
		big.NewInt(10000001):  big.NewInt(2),
		big.NewInt(14999999):  big.NewInt(2),
		big.NewInt(15000000):  big.NewInt(2),
		big.NewInt(15000001):  big.NewInt(3),
		big.NewInt(100000001): big.NewInt(20),
		big.NewInt(123456789): big.NewInt(24),
	}

	for bn, expectedEra := range cases {
		gotEra := GetBlockEra(bn, defaultEraLength)
		if gotEra.Cmp(expectedEra) != 0 {
			t.Errorf("got: %v, want: %v", gotEra, expectedEra)
		}
	}
}

// Use custom era length 2
func TestGetBlockEra2(t *testing.T) {
	cases := map[*big.Int]*big.Int{
		big.NewInt(0):  big.NewInt(0),
		big.NewInt(1):  big.NewInt(0),
		big.NewInt(2):  big.NewInt(0),
		big.NewInt(3):  big.NewInt(1),
		big.NewInt(4):  big.NewInt(1),
		big.NewInt(5):  big.NewInt(2),
		big.NewInt(6):  big.NewInt(2),
		big.NewInt(7):  big.NewInt(3),
		big.NewInt(8):  big.NewInt(3),
		big.NewInt(9):  big.NewInt(4),
		big.NewInt(10): big.NewInt(4),
		big.NewInt(11): big.NewInt(5),
		big.NewInt(12): big.NewInt(5),
	}

	for bn, expectedEra := range cases {
		gotEra := GetBlockEra(bn, big.NewInt(2))
		if gotEra.Cmp(expectedEra) != 0 {
			t.Errorf("got: %v, want: %v", gotEra, expectedEra)
		}
	}
}

func TestGetBlockWinnerRewardByEra(t *testing.T) {
	cases := map[*big.Int]*uint256.Int{
		big.NewInt(0):        MaximumBlockReward,
		big.NewInt(1):        MaximumBlockReward,
		big.NewInt(4999999):  MaximumBlockReward,
		big.NewInt(5000000):  MaximumBlockReward,
		big.NewInt(5000001):  uint256.NewInt(4e+18),
		big.NewInt(9999999):  uint256.NewInt(4e+18),
		big.NewInt(10000000): uint256.NewInt(4e+18),
		big.NewInt(10000001): uint256.NewInt(3.2e+18),
		big.NewInt(14999999): uint256.NewInt(3.2e+18),
		big.NewInt(15000000): uint256.NewInt(3.2e+18),
		big.NewInt(15000001): uint256.NewInt(2.56e+18),
	}

	for bn, expectedReward := range cases {
		gotReward := GetBlockWinnerRewardByEra(GetBlockEra(bn, defaultEraLength), MaximumBlockReward)
		if gotReward.Cmp(expectedReward) != 0 {
			t.Errorf("@ %v, got: %v, want: %v", bn, gotReward, expectedReward)
		}
		if gotReward.Cmp(uint256.NewInt(0)) <= 0 {
			t.Errorf("@ %v, got: %v, want: %v", bn, gotReward, expectedReward)
		}
		if gotReward.Cmp(MaximumBlockReward) > 0 {
			t.Errorf("@ %v, got: %v, want %v", bn, gotReward, expectedReward)
		}
	}
}

func TestGetBlockUncleRewardByEra(t *testing.T) {
	var we1, we2, we3, we4 *uint256.Int = new(uint256.Int), new(uint256.Int), new(uint256.Int), new(uint256.Int)

	// manually divide maxblockreward/32 to compare to got
	we2.Div(GetBlockWinnerRewardByEra(GetBlockEra(big.NewInt(5000001), defaultEraLength), MaximumBlockReward), uint256.NewInt(32))
	we3.Div(GetBlockWinnerRewardByEra(GetBlockEra(big.NewInt(10000001), defaultEraLength), MaximumBlockReward), uint256.NewInt(32))
	we4.Div(GetBlockWinnerRewardByEra(GetBlockEra(big.NewInt(15000001), defaultEraLength), MaximumBlockReward), uint256.NewInt(32))

	cases := map[*big.Int]*uint256.Int{
		big.NewInt(0):        nil,
		big.NewInt(1):        nil,
		big.NewInt(4999999):  nil,
		big.NewInt(5000000):  nil,
		big.NewInt(5000001):  we2,
		big.NewInt(9999999):  we2,
		big.NewInt(10000000): we2,
		big.NewInt(10000001): we3,
		big.NewInt(14999999): we3,
		big.NewInt(15000000): we3,
		big.NewInt(15000001): we4,
	}

	for bn, want := range cases {
		era := GetBlockEra(bn, defaultEraLength)

		var header, uncle *types.Header = &types.Header{}, &types.Header{}
		header.Number = bn

		uncle.Number = big.NewInt(0).Sub(header.Number, big.NewInt(int64(rand.Int31n(int32(7)))))

		got := GetBlockUncleRewardByEra(era, header, uncle, MaximumBlockReward)

		// "Era 1"
		if want == nil {
			we1.Add(uint256.MustFromBig(uncle.Number), big8) // 2,534,998 + 8              = 2,535,006
			we1.Sub(we1, uint256.MustFromBig(header.Number)) // 2,535,006 - 2,534,999        = 7
			we1.Mul(we1, MaximumBlockReward)                 // 7 * 5e+18               = 35e+18
			we1.Div(we1, big8)                               // 35e+18 / 8                            = 7/8 * 5e+18

			if got.Cmp(we1) != 0 {
				t.Errorf("@ %v, want: %v, got: %v", bn, we1, got)
			}
		} else {
			if got.Cmp(want) != 0 {
				t.Errorf("@ %v, want: %v, got: %v", bn, want, got)
			}
		}
	}
}

func TestGetBlockWinnerRewardForUnclesByEra(t *testing.T) {
	// "want era 1", "want era 2", ...
	var we1, we2, we3, we4 *uint256.Int = new(uint256.Int), new(uint256.Int), new(uint256.Int), new(uint256.Int)
	we1.Div(MaximumBlockReward, uint256.NewInt(32))
	we2.Div(GetBlockWinnerRewardByEra(big.NewInt(1), MaximumBlockReward), uint256.NewInt(32))
	we3.Div(GetBlockWinnerRewardByEra(big.NewInt(2), MaximumBlockReward), uint256.NewInt(32))
	we4.Div(GetBlockWinnerRewardByEra(big.NewInt(3), MaximumBlockReward), uint256.NewInt(32))

	cases := map[*big.Int]*uint256.Int{
		big.NewInt(0):        we1,
		big.NewInt(1):        we1,
		big.NewInt(4999999):  we1,
		big.NewInt(5000000):  we1,
		big.NewInt(5000001):  we2,
		big.NewInt(9999999):  we2,
		big.NewInt(10000000): we2,
		big.NewInt(10000001): we3,
		big.NewInt(14999999): we3,
		big.NewInt(15000000): we3,
		big.NewInt(15000001): we4,
	}

	var uncleSingle, uncleDouble []*types.Header = []*types.Header{{}}, []*types.Header{{}, {}}

	for bn, want := range cases {
		// test single uncle
		got := GetBlockWinnerRewardForUnclesByEra(GetBlockEra(bn, defaultEraLength), uncleSingle, MaximumBlockReward)
		if got.Cmp(want) != 0 {
			t.Errorf("@ %v: want: %v, got: %v", bn, want, got)
		}

		// test double uncle
		got = GetBlockWinnerRewardForUnclesByEra(GetBlockEra(bn, defaultEraLength), uncleDouble, MaximumBlockReward)
		dub := new(uint256.Int)
		if got.Cmp(dub.Mul(want, uint256.NewInt(2))) != 0 {
			t.Errorf("@ %v: want: %v, got: %v", bn, want, got)
		}
	}
}

// Integration tests.
//
// There are two kinds of integration tests: accumulating and non-accumulation.
// Accumulating tests check simulated accrual of a
// winner and two uncle accounts over the winnings of many mined blocks.
// If ecip1017 feature is not included in the hardcoded mainnet configuration, it will be temporarily
// included and tested in this test.
// This tests not only reward changes, but summations and state tallies over time.
// Non-accumulating tests check the one-off reward structure at any point
// over the specified era period.
// Currently tested eras are 1, 2, 3, and the beginning of 4.
// Both kinds of tests rely on manual calculations of 'want' account balance state,
// and purposely avoid using existing calculation functions in state_processor.go.
// Check points confirming calculations are at and around the 'boundaries' of forks and eras.
//
// Helpers.

const (
	era1 = 1
	era2 = 2
	era3 = 3
	era4 = 4
)

type expectedRewards map[common.Address]*uint256.Int

func calculateExpectedEraRewards(era *big.Int, numUncles int) expectedRewards {
	wr := new(uint256.Int)
	wur := new(uint256.Int)
	ur := new(uint256.Int)
	uera := era.Int64()
	switch uera {
	case era1:
		wr = Era1WinnerReward
		wur = Era1WinnerUncleReward
		ur = Era1UncleReward
	case era2:
		wr = Era2WinnerReward
		wur = Era2WinnerUncleReward
		ur = Era2UncleReward
	case era3:
		wr = Era3WinnerReward
		wur = Era3WinnerUncleReward
		ur = Era3UncleReward
	case era4:
		wr = Era4WinnerReward
		wur = Era4WinnerUncleReward
		ur = Era4UncleReward
	}
	return expectedRewards{
		WinnerCoinbase: new(uint256.Int).Add(wr, new(uint256.Int).Mul(wur, uint256.NewInt(uint64(numUncles)))),
		Uncle1Coinbase: ur,
		Uncle2Coinbase: ur,
	}
}

// expectedEraFromBlockNumber is similar to GetBlockEra, but it
// returns a 1-indexed version of the number of type expectedEraForTesting
/*func expectedEraFromBlockNumber(i, eralen *big.Int, t *testing.T) int64 {
	e := GetBlockEra(i, eralen)
	// ePlusOne := new(big.Int).Add(e, big.NewInt(1)) // since expectedEraForTesting is not 0-indexed; iota + 1
	ei := ePlusOne.Int64()
	expEra := int(ei)
	if expEra > 4 || expEra < 1 {
		t.Fatalf("Unexpected era value, want 1 < e < 5, got: %d", expEra)
	}
	return int64(expEra)
}*/

type expectedRewardCase struct {
	eraNum  *big.Int
	block   *big.Int
	rewards expectedRewards
}

// String implements stringer interface for expectedRewards
// Useful for logging tests for visual confirmation.
func (r expectedRewards) String() string {
	return fmt.Sprintf("w: %d, u1: %d, u2: %d", r[WinnerCoinbase], r[Uncle1Coinbase], r[Uncle2Coinbase])
}

// String implements stringer interface for expectedRewardCase --
// useful for double-checking test cases with t.Log
// to visually ensure getting all desired test cases.
func (c *expectedRewardCase) String() string {
	return fmt.Sprintf("block=%d era=%d rewards=%s", c.block, c.eraNum, c.rewards)
}

// makeExpectedRewardCasesForConfig makes an array of expectedRewardCases.
// It checks boundary cases for era length and fork numbers.
//
// An example of output:
// ----
//
//	{
//		// mainnet
//		{
//			block:   big.NewInt(2),
//			rewards: calculateExpectedEraRewards(era1, 1),
//		},
//
// ...
//
//		{
//			block:   big.NewInt(20000000),
//			rewards: calculateExpectedEraRewards(era4, 1),
//		},
//	},
func makeExpectedRewardCasesForConfig(c *coregeth.CoreGethChainConfig, numUncles int, t *testing.T) []expectedRewardCase {
	erasToTest := []int64{era1, era2, era3}
	eraLen := defaultEraLength
	ecip1017EraLen := c.ECIP1017EraRounds
	if ecip1017EraLen != nil {
		eraLen = ecip1017EraLen
	}

	var cases []expectedRewardCase
	var boundaryDiffs = []int64{-2, -1, 0, 1, 2}

	// Include trivial initial early block values.
	for _, i := range []*big.Int{big.NewInt(2), big.NewInt(13)} {
		era := GetBlockEra(i, eraLen)
		cases = append(cases, expectedRewardCase{
			eraNum:  era,
			block:   i,
			rewards: calculateExpectedEraRewards(era, numUncles),
		})
	}

	// Test boundaries of era.
	for _, e := range erasToTest {
		for _, d := range boundaryDiffs {
			eb := big.NewInt(e)
			eraBoundary := new(big.Int).Mul(eb, eraLen)
			bn := new(big.Int).Add(eraBoundary, big.NewInt(d))
			if bn.Sign() < 1 {
				t.Fatalf("unexpected 0 or neg block number: %d", bn)
			}
			era := GetBlockEra(bn, eraLen)
			cases = append(cases, expectedRewardCase{
				eraNum:  era,
				block:   bn,
				rewards: calculateExpectedEraRewards(era, numUncles),
			})
		}
	}

	return cases
}

func TestAccumulateRewards(t *testing.T) {
	configs := []*coregeth.CoreGethChainConfig{params.ClassicChainConfig}
	for i, config := range configs {
		cases := [][]expectedRewardCase{}
		cases = append(cases, makeExpectedRewardCasesForConfig(config, 2, t))

		ecip1017ForkBlock := config.ECIP1017FBlock
		if ecip1017ForkBlock == nil {
			t.Fatal("ecip1017ForkBlock is not defined.")
		}

		eraLen := config.ECIP1017EraRounds
		if eraLen == nil {
			t.Error("No era length configured, is required.")
		}

		db := rawdb.NewMemoryDatabase()

		stateDB, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
		if err != nil {
			t.Fatalf("could not open statedb: %v", err)
		}

		var header *types.Header = &types.Header{}
		var uncles []*types.Header = []*types.Header{{}, {}}

		if i == 0 {
			header.Coinbase = common.HexToAddress("000d836201318ec6899a67540690382780743280")
			uncles[0].Coinbase = common.HexToAddress("001762430ea9c3a26e5749afdb70da5f78ddbb8c")
			uncles[1].Coinbase = common.HexToAddress("001d14804b399c6ef80e64576f657660804fec0b")
		} else {
			header.Coinbase = common.HexToAddress("0000000000000000000000000000000000000001")
			uncles[0].Coinbase = common.HexToAddress("0000000000000000000000000000000000000002")
			uncles[1].Coinbase = common.HexToAddress("0000000000000000000000000000000000000003")
		}

		// Manual tallies for reward accumulation.
		totalB := new(uint256.Int)

		blockWinner := *stateDB.GetBalance(header.Coinbase) // start balance. 0
		uncleMiner1 := *stateDB.GetBalance(uncles[0].Coinbase)
		uncleMiner2 := *stateDB.GetBalance(uncles[1].Coinbase)

		totalB.Add(totalB, &blockWinner)
		totalB.Add(totalB, &uncleMiner1)
		totalB.Add(totalB, &uncleMiner2)

		// make sure we are starting clean (everything is 0)
		if !totalB.IsZero() {
			t.Errorf("unexpected: %v", totalB)
		}
		for _, c := range cases[i] {
			bn := c.block
			era := GetBlockEra(bn, eraLen)
			header.Number = bn
			blockReward := ctypes.EthashBlockReward(config, header.Number)
			for i, uncle := range uncles {
				// Randomize uncle numbers with bound ( n-1 <= uncleNum <= n-7 ), where n is current head number
				// See yellowpaper@11.1 for ommer validation reference. I expect n-7 is 6th-generation ommer.
				// Note that ommer nth-generation impacts reward only for "Era 1".
				// 1 + [0..rand..7) == 1 + 0, 1 + 1, ... 1 + 6
				un := new(big.Int).Add(big.NewInt(1), big.NewInt(int64(rand.Int31n(int32(7)))))
				uncle.Number = new(big.Int).Sub(header.Number, un) // n - un

				ur := GetBlockUncleRewardByEra(era, header, uncle, blockReward)
				if i == 0 {
					uncleMiner1.Add(&uncleMiner1, ur)
				}
				if i == 1 {
					uncleMiner2.Add(&uncleMiner2, ur)
				}

				totalB.Add(totalB, ur)
			}

			wr := GetBlockWinnerRewardByEra(era, blockReward)
			wr.Add(wr, GetBlockWinnerRewardForUnclesByEra(era, uncles, blockReward))
			blockWinner.Add(&blockWinner, wr)

			totalB.Add(totalB, &blockWinner)

			AccumulateRewards(config, stateDB, header, uncles)

			// Check balances.
			// t.Logf("config=%d block=%d era=%d w:%d u1:%d u2:%d", i, bn, new(big.Int).Add(era, big.NewInt(1)), blockWinner, uncleMiner1, uncleMiner2)
			if wb := stateDB.GetBalance(header.Coinbase); wb.Cmp(&blockWinner) != 0 {
				t.Errorf("winner balance @ %v, want: %v, got: %v (config: %v)", bn, blockWinner, wb, i)
			}
			if uB0 := stateDB.GetBalance(uncles[0].Coinbase); uncleMiner1.Cmp(uB0) != 0 {
				t.Errorf("uncle1 balance @ %v, want: %v, got: %v (config: %v)", bn, uncleMiner1, uB0, i)
			}
			if uB1 := stateDB.GetBalance(uncles[1].Coinbase); uncleMiner2.Cmp(uB1) != 0 {
				t.Errorf("uncle2 balance @ %v, want: %v, got: %v (config: %v)", bn, uncleMiner2, uB1, i)
			}
		}

		db.Close()
	}
}

func TestGetBlockEra(t *testing.T) {
	blockNum := big.NewInt(11700000)
	eraLength := big.NewInt(5000000)
	era := GetBlockEra(blockNum, eraLength)
	if era.Cmp(big.NewInt(2)) != 0 {
		t.Error("Should return Era 2", "era", era)
	}
	// handle negative blockNum
	blockNum = big.NewInt(-50000)
	era = GetBlockEra(blockNum, eraLength)
	if era.Cmp(big.NewInt(0)) != 0 {
		t.Error("Should return Era 0", "era", era)
	}
	blockNum = big.NewInt(5000001)
	era = GetBlockEra(blockNum, eraLength)
	if era.Cmp(big.NewInt(1)) != 0 {
		t.Error("Should return Era 1", "era", era)
	}
}

func TestGetBlockWinnerRewardByEra2(t *testing.T) {
	baseReward := uint256.NewInt(5000000000000000000)
	era := big.NewInt(0)
	blockReward := GetBlockWinnerRewardByEra(era, baseReward)
	if blockReward.Cmp(uint256.NewInt(5000000000000000000)) != 0 {
		t.Error("Should return blockReward 5000000000000000000", "reward", blockReward)
	}
	era = big.NewInt(1)
	blockReward = GetBlockWinnerRewardByEra(era, baseReward)
	if blockReward.Cmp(uint256.NewInt(4000000000000000000)) != 0 {
		t.Error("Should return blockReward 4000000000000000000", "reward", blockReward)
	}
	era = big.NewInt(2)
	blockReward = GetBlockWinnerRewardByEra(era, baseReward)
	if blockReward.Cmp(uint256.NewInt(3200000000000000000)) != 0 {
		t.Error("Should return blockReward 3200000000000000000", "reward", blockReward)
	}
	era = big.NewInt(3)
	blockReward = GetBlockWinnerRewardByEra(era, baseReward)
	if blockReward.Cmp(uint256.NewInt(2560000000000000000)) != 0 {
		t.Error("Should return blockReward 2560000000000000000", "reward", blockReward)
	}
	era = big.NewInt(4)
	blockReward = GetBlockWinnerRewardByEra(era, baseReward)
	if blockReward.Cmp(uint256.NewInt(2048000000000000000)) != 0 {
		t.Error("Should return blockReward 2048000000000000000", "reward", blockReward)
	}
}

func TestGetRewardForUncle(t *testing.T) {
	baseReward := uint256.NewInt(4000000000000000000)
	era := big.NewInt(0)
	uncleReward := getEraUncleBlockReward(era, baseReward)
	if uncleReward.Cmp(uint256.NewInt(125000000000000000)) != 0 {
		t.Error("Should return uncleReward 125000000000000000", "reward", uncleReward)
	}
	baseReward = uint256.NewInt(3200000000000000000)
	uncleReward = getEraUncleBlockReward(era, baseReward)
	if uncleReward.Cmp(uint256.NewInt(100000000000000000)) != 0 {
		t.Error("Should return uncleReward 100000000000000000", "reward", uncleReward)
	}
	baseReward = uint256.NewInt(2560000000000000000)
	uncleReward = getEraUncleBlockReward(era, baseReward)
	if uncleReward.Cmp(uint256.NewInt(80000000000000000)) != 0 {
		t.Error("Should return uncleReward 80000000000000000", "reward", uncleReward)
	}
	baseReward = uint256.NewInt(2048000000000000000)
	uncleReward = getEraUncleBlockReward(era, baseReward)
	if uncleReward.Cmp(uint256.NewInt(64000000000000000)) != 0 {
		t.Error("Should return uncleReward 64000000000000000", "reward", uncleReward)
	}
}
