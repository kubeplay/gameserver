package types

import "fmt"

type Object interface {
	GetObjectKind() string
	GetObjectMeta() *Metadata
	New() Object
}

// TypeMeta describes an individual object in an API response or request
// with strings representing the type of the object and its API schema version.
// Structures that are versioned or persisted should inline TypeMeta.
type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	Kind string `json:"kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// APIVersion string `json:"apiVersion"`
}

type ListMeta struct{}

type Metadata struct {
	Name        string            `json:"name"`
	UID         string            `json:"uid"`
	CreatedAt   string            `json:"createdAt"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (m *ListMeta) GetObjectMeta() *Metadata {
	return nil
}

func (m *Metadata) GetObjectMeta() *Metadata {
	return m
}

func (t *TypeMeta) GetObjectKind() string {
	return t.Kind
}

func (o *Game) New() Object          { return &Game{} }
func (o *GameList) New() Object      { return &GameList{} }
func (o *Challenge) New() Object     { return &Challenge{} }
func (o *ChallengeList) New() Object { return &ChallengeList{} }
func (o *Player) New() Object        { return &Player{} }
func (o *PlayerList) New() Object    { return &PlayerList{} }
func (o *Event) New() Object         { return &Event{} }
func (o *EventList) New() Object     { return &EventList{} }

func (c *PlayerClaims) Username() string {
	return fmt.Sprintf("github|%s", c.Login)
}
