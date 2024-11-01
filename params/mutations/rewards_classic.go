// Copyright 2019 The multi-geth Authors
// This file is part of the multi-geth library.
//
// The multi-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The multi-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the multi-geth library. If not, see <http://www.gnu.org/licenses/>.
package mutations

import (
	"math/big"

	"github.com/shudolab/core-geth/core/types"
	"github.com/shudolab/core-geth/params/types/ctypes"
	"github.com/shudolab/core-geth/params/vars"
	"github.com/holiman/uint256"
)

func ecip1017BlockReward(config ctypes.ChainConfigurator, header *types.Header, uncles []*types.Header) (*uint256.Int, []*uint256.Int) {
	blockReward := vars.FrontierBlockReward

	// Ensure value 'era' is configured.
	eraLen := config.GetEthashECIP1017EraRounds()
	era := GetBlockEra(header.Number, new(big.Int).SetUint64(*eraLen))
	wr := GetBlockWinnerRewardByEra(era, blockReward)                    // wr "winner reward". 5, 4, 3.2, 2.56, ...
	wurs := GetBlockWinnerRewardForUnclesByEra(era, uncles, blockReward) // wurs "winner uncle rewards"
	wr.Add(wr, wurs)

	// Reward uncle miners.
	uncleRewards := make([]*uint256.Int, len(uncles))
	for i, uncle := range uncles {
		ur := GetBlockUncleRewardByEra(era, header, uncle, blockReward)
		uncleRewards[i] = ur
	}

	return wr, uncleRewards
}

// GetBlockEra gets which "Era" a given block is within, given an era length (ecip-1017 has era=5,000,000 blocks)
// Returns a zero-index era number, so "Era 1": 0, "Era 2": 1, "Era 3": 2 ...
func GetBlockEra(blockNum, eraLength *big.Int) *big.Int {
	// If genesis block or impossible negative-numbered block, return zero-val.
	if blockNum.Sign() < 1 {
		return new(big.Int)
	}

	remainder := big.NewInt(0).Mod(big.NewInt(0).Sub(blockNum, big.NewInt(1)), eraLength)
	base := big.NewInt(0).Sub(blockNum, remainder)

	d := big.NewInt(0).Div(base, eraLength)
	dremainder := big.NewInt(0).Mod(d, big.NewInt(1))

	return new(big.Int).Sub(d, dremainder)
}
