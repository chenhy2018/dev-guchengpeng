package mockio

import (
	"qbox.us/sstore"
)

const KeyHintESTest = 139
const KeyHintMockRS = 133
const KeyHintMockFS = 113
const KeyHintFSTest = 103
const KeyHintRSTest = 105
const KeyHintRsPlusTest = 108
const KeyHintRsCacheTest = 111
const KeyHintIOTest = 104

var KeyMockRS = []byte("qbox.mockrs")
var KeyMockFS = []byte("qbox.mockfs")
var KeyFSTest = []byte("qbox.fs.test")
var KeyESTest = []byte("qbox.es.test")
var KeyRSTest = []byte("qbox.rs.test")
var KeyRsPlusTest = []byte("qbox.rsplus.test")
var KeyRsCacheTest = []byte("qbox.rscache.test")
var KeyIOTest = []byte("qbox.io.test")

var KeyFinder = sstore.SimpleKeyFinder(map[uint32][]byte{
	KeyHintMockRS:      KeyMockRS,
	KeyHintMockFS:      KeyMockFS,
	KeyHintFSTest:      KeyFSTest,
	KeyHintESTest:      KeyESTest,
	KeyHintRSTest:      KeyRSTest,
	KeyHintRsPlusTest:  KeyRsPlusTest,
	KeyHintRsCacheTest: KeyRsCacheTest,
	KeyHintIOTest:      KeyIOTest,
})
