package perms

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestWeb_AddPConf(t *testing.T) {
	web := NewWeb()

	pconf := MustParsePConf([]byte(`{
        "groups": {
			"default": {
				"nodes": [
					"profile.use"
				]
			},
            "admin": {
                "nodes": [
                    "billing.*"
                ]
            }
        },
        "users": {
            "ammar": {
                "groups": [
                    "admin"
                ],
                "nodes": [
                    "projects.backend.create"
                ]
            }
        }
    }`))

	if err := web.AddPConf(pconf); err != nil {
		t.Errorf("err while adding pconf: %v", err)
	}

	t.Run("check_user", func(t *testing.T) {
		u := web.GetUser("ammar")

		if u == nil {
			t.Errorf("user is nil")
		}

		if !reflect.DeepEqual(u.Groups, []string{"admin"}) {
			t.Errorf("u.Groups[0] should be 'admin' but is %v", u.Groups)
		}

		if !reflect.DeepEqual(u.Nodes, Nodes{MustParseNode("projects.backend.create")}) {
			t.Errorf("u.Nodes is %v", u.Nodes)
		}
	})

	t.Run("check_group", func(t *testing.T) {
		g := web.GetGroup("admin")

		if g == nil {
			t.Errorf("group is nil")
		}

		if !reflect.DeepEqual(g.Nodes, Nodes{MustParseNode("billing.*")}) {
			t.Errorf("g.Nodes is %v", g.Nodes)
		}

		if !reflect.DeepEqual(g.Name, "admin") {
			t.Errorf("g.Name is %v", g.Name)
		}
	})

	t.Run("check_perms", func(t *testing.T) {
		if !web.CheckUserHasPermission("ammar", MustParseNode("billing.manage")) {
			t.Errorf("ammar should have billing.manage")
		}
		if !web.CheckUserHasPermission("ammar", MustParseNode("profile.use")) {
			t.Errorf("ammar should have profile.use")
		}
	})
}

func TestWeb_PrettyDump(t *testing.T) {
	if !testing.Verbose() {
		return
	}

	web := NewWeb()

	web.AddGroup(NewGroup("admin"))
	web.AddGroup(NewGroup("moderator"))

	web.AddUser(&User{
		Name:   "ammar",
		Groups: []string{"admin", "moderator"},
		Nodes:  []Node{MustParseNode("testing.t")},
	})

	web.PrettyDump(os.Stdout)
}

func TestWeb_MasterPConf(t *testing.T) {
	if !testing.Verbose() {
		return
	}

	web := NewWeb()

	web.AddGroup(&Group{
		Name:  "admin",
		Nodes: MustParseNodes(strings.NewReader("testing.create testing.poop testing.weewee")),
	})

	web.AddUser(&User{
		Name:  "ahmad",
		Nodes: MustParseNodes(strings.NewReader("ammar.edit ammar.poo ammar.pee")),
	})

	raw, err := web.MasterPConf().PrettyMarshal()
	if err != nil {
		t.Fatalf("Failed to generate master pconf: %v", err)
	}

	fmt.Printf("MasterPconf\n %s\n", raw)
}

func TestWeb_JSON(t *testing.T) {
	w := NewWeb()
	expected := []byte(`{"groups":{},"users":{"ammar":{"groups":["slave","coder"],"nodes":["everything.*","-nothing"]}}}`)

	t.Run("MarshalJSON", func(t *testing.T) {
		_ = json.Marshaler(w)
		w.AddUser(&User{
			Name:   "ammar",
			Groups: []string{"slave", "coder"},
			Nodes:  MustParseNodes(strings.NewReader("everything.* -nothing")),
		})
		js, err := w.MarshalJSON()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}
		if !reflect.DeepEqual(js, expected) {
			t.Errorf("js == %s\n", js)
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		_ = json.Unmarshaler(w)
		w.Reset()

		//this thing should be deleted
		w.AddGroup(&Group{Name: "admin"})

		if err := w.UnmarshalJSON(expected); err != nil {
			t.Errorf("Failed to unmarshal: %v", err)
		}
	})
}
