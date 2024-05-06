package utils

func GetDirectionIdxByTargetPosition(currentX, currentY, targetX, targetY float64) int {
	if targetX < currentX {
		if targetY < currentY {
			return 2
		} else {
			return 1
		}
	} else {
		if targetY < currentY {
			return 3
		} else {
			return 0
		}
	}
}
