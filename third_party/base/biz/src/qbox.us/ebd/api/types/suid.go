package types

func DecodeSuid(suid uint64) (sid uint64, idx int) {
	sid, idx = suid/(N+M), int(suid%(N+M))
	return
}

func EncodeSuid(sid uint64, idx int) (suid uint64) {
	suid = sid*(N+M) + uint64(idx)
	return
}
