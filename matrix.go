package main

import "math/big"

func MultiplyMatrices(a, b [3][3]*big.Int) ([3][3]*big.Int, [9]*big.Int) {
	var result [3][3]*big.Int
	var singleArray [9]*big.Int
	position := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			result[i][j] = new(big.Int)
			for k := 0; k < 3; k++ {
				temp := new(big.Int).Mul(a[i][k], b[k][j])
				result[i][j].Add(result[i][j], temp)
			}
			singleArray[position] = result[i][j]
			position += 1
		}
	}
	return result, singleArray
}
