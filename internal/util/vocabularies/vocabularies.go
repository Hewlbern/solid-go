package vocabularies

import "solid-go/internal/util/n3"

// ACL contains Web Access Control vocabulary terms
var ACL = struct {
	Agent              n3.Term
	AgentClass         n3.Term
	AgentGroup         n3.Term
	AuthenticatedAgent n3.Term
}{
	Agent:              &BasicTerm{value: "http://www.w3.org/ns/auth/acl#agent"},
	AgentClass:         &BasicTerm{value: "http://www.w3.org/ns/auth/acl#agentClass"},
	AgentGroup:         &BasicTerm{value: "http://www.w3.org/ns/auth/acl#agentGroup"},
	AuthenticatedAgent: &BasicTerm{value: "http://www.w3.org/ns/auth/acl#AuthenticatedAgent"},
}

// FOAF contains Friend of a Friend vocabulary terms
var FOAF = struct {
	Agent n3.Term
}{
	Agent: &BasicTerm{value: "http://xmlns.com/foaf/0.1/Agent"},
}

// VCARD contains vCard vocabulary terms
var VCARD = struct {
	HasMember n3.Term
}{
	HasMember: &BasicTerm{value: "http://www.w3.org/2006/vcard/ns#hasMember"},
}

// BasicTerm is a simple implementation of the Term interface
type BasicTerm struct {
	value string
}

// Value implements n3.Term.Value
func (t *BasicTerm) Value() string {
	return t.value
}
