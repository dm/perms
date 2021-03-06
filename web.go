package perms

import (
	"bufio"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

//Web is an isolated permissions system
type Web struct {
	groups map[string]*Group
	users  map[string]*User
}

//NewWeb returns an instantiated web
func NewWeb() *Web {
	w := &Web{}
	w.Reset()
	return w
}

//Reset resets the state of w
func (w *Web) Reset() {
	w.groups = make(map[string]*Group, 20)
	w.users = make(map[string]*User, 20)
}

//AddPConf adds a PConf to the web
func (w *Web) AddPConf(p *PConf) error {
	for name, unprocessedGroup := range p.Groups {
		group := NewGroup(name)
		w.groups[name] = group

		for _, nodeStr := range unprocessedGroup.Nodes {
			node, err := ParseNode(nodeStr)
			if err != nil {
				return errors.Wrapf(err, "failed to parse node %q", nodeStr)
			}
			group.Nodes = append(group.Nodes, node)
		}
		group.Parents = unprocessedGroup.Parents
	}
	for name, unprocessedUser := range p.Users {
		user := NewUser(name)
		w.users[name] = user

		for _, nodeStr := range unprocessedUser.Nodes {
			node, err := ParseNode(nodeStr)
			if err != nil {
				return errors.Wrapf(err, "failed to parse node %q", nodeStr)
			}
			user.Nodes = append(user.Nodes, node)
		}
		user.Groups = unprocessedUser.Groups
	}
	return nil
}

//AddUser adds a user to the web
func (w *Web) AddUser(u *User) {
	if u.Groups == nil {
		u.Groups = []string{}
	}
	if u.Nodes == nil {
		u.Nodes = Nodes{}
	}
	w.users[u.Name] = u
}

//GetUser returns a user with name
func (w *Web) GetUser(name string) *User {
	return w.users[name]
}

//DelUser deletes a user
func (w *Web) DelUser(name string) {
	delete(w.users, name)
}

//AddGroup adds a group to the web.
//It instantiates nil values
func (w *Web) AddGroup(g *Group) {
	if g.Nodes == nil {
		g.Nodes = Nodes{}
	}
	if g.Parents == nil {
		g.Parents = []string{}
	}
	w.groups[g.Name] = g
}

//GetGroup gets a group. It returns nil if no group of name exists in web
func (w *Web) GetGroup(name string) *Group {
	return w.groups[name]
}

//DelGroup deletes a group from the web
func (w *Web) DelGroup(name string) {
	delete(w.groups, name)
}

//CheckUserHasPermission checks is a user has a permission.
//It is negation aware.
func (w *Web) CheckUserHasPermission(name string, check Node) bool {
	user := w.users[name]

	if user == nil {
		return false
	}

	//Check user's direct permissions first
	matched, negated := user.Nodes.Check(check)

	if negated {
		return false
	} else if matched {
		return true
	}

	//Matched has to be false here

	if defaultGroup, exists := w.groups["default"]; exists {
		thisMatched, negated := defaultGroup.Nodes.Check(check)
		if negated {
			return false
		}
		if thisMatched {
			matched = true
		}
	}
	//fmt.Printf("Node %v Matched %v\n", check, matched)

	//Check user's groups for permissions
	for _, groupName := range user.Groups {
		group := w.groups[groupName]
		if group == nil {
			continue
		}
		thisMatched, negated := group.Nodes.Check(check)
		if negated {
			//If it is ever negated now we know they don't have the node
			return false
		}
		if thisMatched {
			matched = true
		}
	}
	//fmt.Printf("Node %v Matched %v\n", check, matched)

	return matched
}

//MasterPConf generates a serialized master pconf
func (w *Web) MasterPConf() (pconf *PConf) {
	pc := newPConf()
	for name, group := range w.groups {
		pc.Groups[name] = pconfGroup{
			Parents: group.Parents,
			Nodes:   group.Nodes.Strings(),
		}
	}
	for name, user := range w.users {
		pc.Users[name] = pconfUser{
			Groups: user.Groups,
			Nodes:  user.Nodes.Strings(),
		}
	}
	return pc
}

//MarshalJSON marshal's w into valid json
func (w *Web) MarshalJSON() ([]byte, error) {
	return w.MasterPConf().Marshal()
}

//UnmarshalJSON unmarshals json in b into w.
//UnmarshalJSON resets the state of w.
func (w *Web) UnmarshalJSON(b []byte) error {
	pconf, err := ParsePConf(b)
	if err != nil {
		errors.Wrap(err, "failed to parse pconf")
	}
	return errors.Wrap(w.AddPConf(pconf), "failed to add pconf")
}

//PrettyDump outputs a pretty version of the web to a writer
func (w *Web) PrettyDump(wr io.Writer) error {
	bufw := bufio.NewWriter(wr)

	fmt.Fprintf(bufw, "%v Groups\n", len(w.groups))
	for k, v := range w.groups {
		fmt.Fprintf(bufw, "   %v:\n", k)
		fmt.Fprintf(bufw, "      %v Parents:\n", len(v.Parents))
		for _, parent := range v.Parents {
			fmt.Fprintf(bufw, "         %v\n", parent)
		}
		fmt.Fprintf(bufw, "      %v Nodes:\n", len(v.Nodes))
		for _, node := range v.Nodes {
			fmt.Fprintf(bufw, "         %v\n", node)
		}
	}

	fmt.Fprintf(bufw, "%v Users\n", len(w.users))
	for k, v := range w.users {
		fmt.Fprintf(bufw, "   %v:\n", k)
		fmt.Fprintf(bufw, "      %v Groups:\n", len(v.Groups))
		for _, group := range v.Groups {
			fmt.Fprintf(bufw, "         %v\n", group)
		}
		fmt.Fprintf(bufw, "      %v Nodes:\n", len(v.Nodes))
		for _, node := range v.Nodes {
			fmt.Fprintf(bufw, "         %v\n", node)
		}
	}

	return bufw.Flush()
}
