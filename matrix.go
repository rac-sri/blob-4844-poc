package main

func MultiplyMatrices(a, b [3][3]int) [3][3]int {
	var result [3][3]int
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			result[i][j] = 0
			for k := 0; k < 3; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}
