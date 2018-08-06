package types

const ZeroBlockPsect = ^uint64(0)

func DecodePsect(psect uint64) (diskId, iSector uint32) {
	diskId, iSector = uint32(psect>>32), uint32(psect)
	return
}

func EncodePsect(diskId, iSector uint32) (psect uint64) {
	psect = uint64(diskId)<<32 ^ uint64(iSector)
	return
}
