package perms

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseNodes(t *testing.T) {
	type args struct {
		raw []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Nodes
		wantErr bool
	}{
		{
			"simple",
			args{raw: []byte("project.test project.build")},
			Nodes{
				MustParseNode("project.test"),
				MustParseNode("project.build"),
			},
			false,
		},
		{
			"simple",
			args{raw: []byte("project.test\nproject.build")},
			Nodes{
				MustParseNode("project.test"),
				MustParseNode("project.build"),
			},
			false,
		},
		{
			"whitepower",
			args{raw: []byte("   project.test  \nproject.build ")},
			Nodes{
				MustParseNode("project.test"),
				MustParseNode("project.build"),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNodes(bytes.NewReader(tt.args.raw))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodes_String(t *testing.T) {
	tests := []struct {
		name string
		ns   Nodes
		want string
	}{
		{
			"simple",
			Nodes{
				MustParseNode("project.build"),
				MustParseNode("project.test"),
			},
			"project.build\nproject.test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ns.String(); got != tt.want {
				t.Errorf("Nodes.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodes_Check(t *testing.T) {
	u := NewUser("ammar")

	u.Nodes = Nodes{MustParseNode("projects.test"), MustParseNode("projects.build"), MustParseNode("projects.chat")}

	matched, negated := u.Nodes.Check(MustParseNode("projects.test"))

	if negated {
		t.Fatalf("negated should be false")
	}

	if !matched {
		t.Fatalf("matched should be true")
	}

	u.Nodes = append(u.Nodes, MustParseNode("-projects.test"))

	matched, negated = u.Nodes.Check(MustParseNode("projects.test"))

	if !negated {
		t.Fatalf("negated should be true")
	}

	if !matched {
		t.Fatalf("matched should be true")
	}

}

func TestNodes_SQL(t *testing.T) {
	input := "test.hello\ntest2.hello"
	var nodes Nodes
	if err := nodes.Scan(input); err != nil {
		t.Fatalf("scanning nodes should not return an error; %s", nodes)
	}

	matched, negated := nodes.Check(MustParseNode("test.hello"))
	if !matched {
		t.Fatalf("matched should be true")
	}
	if negated {
		t.Fatalf("negated should be false")
	}

	value, err := nodes.Value()
	if err != nil {
		t.Fatalf("Value should not return an error")
	}
	vstr := value.(string)
	if vstr != input {
		t.Fatalf("Value should return the same string as Scan input, got %s", vstr)
	}
}
