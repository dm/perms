package perms

import (
	"reflect"
	"testing"
)

func TestParseNode(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		want    Node
		wantErr bool
	}{
		{"simple", args{"projects.manage"}, Node{Parts: []string{"projects", "manage"}}, false},
		{"simple", args{"projects.manage.*"}, Node{Parts: []string{"projects", "manage", "*"}}, false},
		{"negate", args{"-projects.manage.*"}, Node{Parts: []string{"projects", "manage", "*"}, Negate: true}, false},
		{"whitespace", args{"- projects.manage.*"}, Node{}, true},
		{"empty", args{"..*"}, Node{}, true},
		{"empty", args{""}, Node{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNode(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkParseNode(b *testing.B) {
	permString := "projects.webserver.use"

	for i := 0; i < b.N; i++ {
		ParseNode(permString)
	}
}

func TestNode_Match(t *testing.T) {
	type args struct {
		check Node
	}
	tests := []struct {
		name string
		n    Node
		args args
		want bool
	}{
		{"simple", Node{Parts: []string{"projects", "webserver"}}, args{Node{Parts: []string{"projects", "webserver"}}}, true},
		{"simple", Node{Parts: []string{"projects", "webserver"}}, args{Node{Parts: []string{"projects", "frontend"}}}, false},

		{"wildcard", Node{Parts: []string{"projects", "*"}}, args{Node{Parts: []string{"projects", "frontend"}}}, true},
		{"wildcard", Node{Parts: []string{"projects", "*"}}, args{Node{Parts: []string{"billing", "frontend"}}}, false},

		{"middle_wildcard", Node{Parts: []string{"projects", "*", "chat"}}, args{Node{Parts: []string{"projects", "test"}}}, false},
		{"middle_wildcard", Node{Parts: []string{"projects", "*", "chat"}}, args{Node{Parts: []string{"projects", "test", "test"}}}, false},
		{"middle_wildcard", Node{Parts: []string{"projects", "*", "chat"}}, args{Node{Parts: []string{"projects", "test", "chat"}}}, true},

		{"supernode", Node{Parts: []string{"*"}}, args{Node{Parts: []string{"projects", "test", "chat"}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Match(tt.args.check); got != tt.want {
				t.Errorf("Node.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkNode_Match(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		node := MustParseNode("projects.backend.use")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			node.Match(node)
		}
	})
	b.Run("wildcard", func(b *testing.B) {
		node := MustParseNode("*")
		checkNode := MustParseNode("projects.backend.use")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			node.Match(checkNode)
		}
	})
	b.Run("middle_wildcard", func(b *testing.B) {
		node := MustParseNode("projects.*.use")
		checkNode := MustParseNode("projects.backend.use")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			node.Match(checkNode)
		}
	})
}

func TestNode_String(t *testing.T) {
	type fields struct {
		Namespaces []string
		Negate     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simple", fields{Namespaces: []string{"projects", "backend"}}, "projects.backend"},
		{"supernode", fields{Namespaces: []string{"*"}}, "*"},
		{"negate", fields{Namespaces: []string{"billing", "*"}, Negate: true}, "-billing.*"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Node{
				Parts:  tt.fields.Namespaces,
				Negate: tt.fields.Negate,
			}
			if got := n.String(); got != tt.want {
				t.Errorf("Node.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkNode_String(b *testing.B) {
	node := MustParseNode("-billing.credit_cards.view")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node.String()
	}
}
