package main

func Log2(k uint64) int {
	i := int(k % uint64(len(tbl_Log2)))
	return tbl_Log2[i]
}

var tbl_Log2 = [67]int{
	-1,
	0,
	1,
	39,
	2,
	15,
	40,
	23,
	3,
	12,
	16,
	59,
	41,
	19,
	24,
	54,
	4,
	-1,
	13,
	10,
	17,
	62,
	60,
	28,
	42,
	30,
	20,
	51,
	25,
	44,
	55,
	47,
	5,
	32,
	-1,
	38,
	14,
	22,
	11,
	58,
	18,
	53,
	63,
	9,
	61,
	27,
	29,
	50,
	43,
	46,
	31,
	37,
	21,
	57,
	52,
	8,
	26,
	49,
	45,
	36,
	56,
	7,
	48,
	35,
	6,
	34,
	33,
}
