package smt

// getBitFromMSB gets the bit at an offset from the most significant bit
func getBitFromMSB(data []byte, position int) int {
	if int(data[position/8])&(1<<(8-1-uint(position)%8)) > 0 {
		return 1
	}
	return 0
}

// setBitFromMSB sets the bit at an offset from the most significant bit
func setBitFromMSB(data []byte, position int) {
	n := int(data[position/8])
	n |= 1 << (8 - 1 - uint(position)%8)
	data[position/8] = byte(n)
}

//countSetBits counts the bits from MSB.
func countSetBits(data []byte) int {
	count := 0
	for i := 0; i < len(data)*8; i++ {
		if getBitFromMSB(data, i) == 1 {
			count++
		}
	}
	return count
}

//countCommonPrefix counts the number of common prefix
func countCommonPrefix(data1 []byte, data2 []byte) int {
	count := 0
	for i := 0; i < len(data1)*8; i++ {
		if getBitFromMSB(data1, i) == getBitFromMSB(data2, i) {
			count++
		} else {
			break
		}
	}
	return count
}

//emptyBytes return empty slice of bytes of length
func emptyBytes(length int) []byte {
	b := make([]byte, length)
	return b
}

//reverseByteSlices reverses the byte slice
func reverseByteSlices(slices [][]byte) [][]byte {
	for left, right := 0, len(slices)-1; left < right; left, right = left+1, right-1 {
		slices[left], slices[right] = slices[right], slices[left]
	}

	return slices
}
