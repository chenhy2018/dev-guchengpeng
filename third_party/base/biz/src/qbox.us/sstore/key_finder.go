package sstore

type SimpleKeyFinder map[uint32][]byte

func NewSimpleKeyFinder() SimpleKeyFinder {
	return make(SimpleKeyFinder)
}

func (keys SimpleKeyFinder) Find(keyHint uint32) []byte {
	if key, ok := keys[keyHint]; ok {
		return key
	}
	return nil
}

func (keys SimpleKeyFinder) Add(keyHint uint32, key []byte) {
	keys[keyHint] = key
}
