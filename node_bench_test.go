package perms

import (
	"fmt"
	"testing"
)

func BenchmarkNode_Parse(b *testing.B) {
	bench := func(node string) {
		b.Run(node, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				MustParseNode(node)
			}
		})
	}

	bench("pretty.generic.node")
	bench("-negated.generic.node")
	bench("wildcard.*.node")
	bench("l.e.n.g.t.h.y")
}

func BenchmarkNode_Match(b *testing.B) {
	bench := func(node string, check string) {
		b.Run(fmt.Sprintf("match %v in %v", check, node), func(b *testing.B) {
			node := MustParseNode(node)
			check := MustParseNode(check)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.Match(check)
			}
		})
	}

	bench("webserver.use", "webserver.use")
	bench("*", "webserver.use")
	bench("webserver.*.use", "webserver.fun.use")
}

func BenchmarkNode_String(b *testing.B) {
	node := MustParseNode("-billing.credit_cards.view")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node.String()
	}
}
