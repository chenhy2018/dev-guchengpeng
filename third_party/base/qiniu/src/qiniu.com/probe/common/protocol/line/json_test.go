package line

import "testing"

func BenchmarkMarshalJson(b *testing.B) {
	m := map[string]interface{}{
		"aa": 100,
		"bb": true,
		"cc": 123.321,
		"dd": `sdfw23" ,sfs=sdf`,
		"ee": []int{1, 2, 3},
		"ff": map[string]interface{}{
			"zz": `sfwf3fw `,
		},
	}
	for i := 0; i < b.N; i++ {
		_ = MarshalJson(m)
	}
}
