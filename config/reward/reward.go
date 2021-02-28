package reward

import (
	"math"
	"pandora-pay/config"
)

func GetRewardAt(blockHeight uint64) (reward uint64) {

	cycle := int(math.Floor(float64(blockHeight) / blocksPerCycle()))

	reward = 4000 / (1 << cycle)

	if reward < 1 {
		reward = 0
	}

	return config.ConvertToUnits(reward)
}

// halving every year
func blocksPerCycle() float64 {
	return 1 * 365.25 * 24 * 60 * 60 / float64(config.BLOCK_TIME)
}
