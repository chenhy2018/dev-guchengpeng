package types

type Uint32Slice []uint32

func (p *Uint32Slice) RemoveDuplicates() {
	var ret Uint32Slice

	pTmp := *p

	if len(pTmp) == 0 {
		return
	}

	m := map[uint32]bool{}

	for _, v := range pTmp {
		m[v] = true
	}

	for k, _ := range m {
		ret = append(ret, k)
	}

	*p = ret
}
