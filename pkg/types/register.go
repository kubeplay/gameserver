package types

import (
	"encoding/json"
	"fmt"
)

type Kind string

const (
	ChallengeKind = "Challenge"
	GameKind      = "Game"
	EventKind     = "Event"
	PolicyKind    = "Policy"
)

var RegisteredTypes = []Object{
	&Game{TypeMeta: TypeMeta{Kind: GameKind}},
	&Challenge{TypeMeta: TypeMeta{Kind: ChallengeKind}},
	&Event{TypeMeta: TypeMeta{Kind: EventKind}},
	&Policy{TypeMeta: TypeMeta{Kind: PolicyKind}},
}

func Decode(meta *TypeMeta, payload []byte) (Object, error) {
	var result Object
	for _, obj := range RegisteredTypes {
		if meta.GetObjectKind() == obj.GetObjectKind() {
			result = obj.New()
			if err := json.Unmarshal(payload, result); err != nil {
				return nil, fmt.Errorf("failed decoding payload %v", err)
			}
		}
	}
	return result, nil
}
