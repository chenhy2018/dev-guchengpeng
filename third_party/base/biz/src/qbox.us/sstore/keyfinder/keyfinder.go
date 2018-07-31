package keyfinder

import (
	"qbox.us/sstore"
)

const KeyHintESTest = 139
const KeyHintMockFS = 113
const KeyHintFSTest = 103
const KeyHintIOTest = 104
const KeyHintRSTest = 105
const KeyHintPubTest = 106

var KeyMockFS = []byte("qbox.mockfs")
var KeyFSTest = []byte("qbox.fs.test")
var KeyIOTest = []byte("qbox.io.test")
var KeyESTest = []byte("qbox.es.test")
var KeyRSTest = []byte("qbox.rs.test")
var KeyPubTest = []byte("qbox.pub.test")

var KeyFinder = sstore.SimpleKeyFinder(map[uint32][]byte{
	KeyHintMockFS:  KeyMockFS,
	KeyHintFSTest:  KeyFSTest,
	KeyHintIOTest:  KeyIOTest,
	KeyHintESTest:  KeyESTest,
	KeyHintRSTest:  KeyRSTest,
	KeyHintPubTest: KeyPubTest,
})
