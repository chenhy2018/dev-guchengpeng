package qetag

import (
	"io"
	"testing"
)

type sha1Test struct {
	out string
	in  string
}

var golden = []sha1Test{
	{"Fto5o-5ea0sNMlW_75VgGJCv2AcJ", ""},
	{"Fob35Df6paf84V0d3Lnq6uo3dme4", "a"},
	{"FtojYU4CRpoNfHvRvatcnEdLGQTc", "ab"},
	{"FqmZPjZHBoFquj4lcXhQwmyc0Nid", "abc"},
	{"FoH-i_6HV2w-yyJCb45XhHOCkXrP", "abcd"},
	{"FgPebFcL_iS_wyjM18pGt26tr0M0", "abcde"},
	{"Fh-KwQ8jxbW8EWe9qEuDPlwFenfS", "abcdef"},
	{"Fi-14TQZ_IkkaGXnoyT0duxiTodA", "abcdefg"},
	{"FkJa8SoHQ1ArMi6ToBW8-GjjJNVq", "abcdefgh"},
	{"FsY7GfHkyLX3ayXEm4uH9X2OSHKh", "abcdefghi"},
	{"FtaMGaCjRbfqt41eEemRwCbsYNtj", "abcdefghij"},
	{"Fuv4Hdy-W_E6qr3E1lNU_fIETzin", "Discard medicine more than two years old."},
	{"FuXeoJOS3YhspjUxqqAFcdwHVUu2", "He who has a shady past knows that nice guys finish last."},
	{"FkWYj3I0RnuU4-lJRDTJbuNgnY-P", "I wouldn't marry him with a ten foot pole."},
	{"FlXe4DfrdGDVppLRzhEzCyYOQMmI", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{"Fre8X7kQgMfea1guooH4o5bXwK7o", "The days of the digital watch are numbered.  -Tom Stoppard"},
	{"FsOu2TWPfHf1I6_oYTXwa5WzmZeX", "Nepal premier won't resign."},
	{"Fm4p0wK_bjpeQwX_MY2YMZfWkGu5", "For every action there is an equal and opposite government program."},
	{"Fll_alQAEPlMFdcYBqmaLIcQ50e9", "His money is twice tainted: 'taint yours and 'taint mine."},
	{"FmhZczslkKigkc7PUAhv68XO7x6A", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{"FlFLJjDsCJuK7hh5X8DPH0hgzayt", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{"FsXKDUp7Znb8eqcsqkHMPV31Z-1p", "size:  a.out:  bad magic"},
	{"FnTFH6mgTq3Iwbvqp_xEL4NLkKAK", "The major problem is with sendmail.  -Mark Horton"},
	{"FgtMTOX1LDrSghhSqNwAIX-hi4tm", "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{"Fjrnk33XkDFb6w9IMw6GQiN8YVUK", "If the enemy is within range, then so are you."},
	{"FkEKKylt-SuaR0ErEygd-PgwqfRL", "It's well we cannot hear the screams/That we create in others' dreams."},
	{"FoQefIXKGtzdvdAYfxKJrLXGQvf1", "You remind me of a TV show, but that's all right: I watch it anyway."},
	{"FhYxc7gl0DuVJgE3ayUhLfZnY-Hb", "C is as portable as Stonehedge!!"},
	{"FjKwN38mh-uI4iEG8TPFhqsxTVJ5", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{"FgiFqvmbVpVC_RZfpE4yJxj0qYTg", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{"FmYn1pBNcUILC_OIarYpYjU4aJ9F", "How can you write a big system without C++?  -Paul Glick"},
}

func TestGolden(t *testing.T) {
	for i := 0; i < len(golden); i++ {
		g := golden[i]
		c := New()
		for j := 0; j < 3; j++ {
			if j < 2 {
				io.WriteString(c, g.in)
			} else {
				io.WriteString(c, g.in[0:len(g.in)/2])
				// c.Sum(nil)
				io.WriteString(c, g.in[len(g.in)/2:])
			}
			s := Qetag(c.Sum(nil)).String()
			if s != g.out {
				t.Fatalf("sha1[%d](%s) = %s want %s", j, g.in, s, g.out)
			}
			s = Sum([]byte(g.in)).String()
			if s != g.out {
				t.Fatalf("sha1[%d](%s) = %s want %s", j, g.in, s, g.out)
			}
			c.Reset()
		}
	}
}
