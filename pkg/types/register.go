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
	PlayerKind    = "Player"
)

var RegisteredTypes = []Object{
	&Game{TypeMeta: TypeMeta{Kind: GameKind}},
	&Challenge{TypeMeta: TypeMeta{Kind: ChallengeKind}},
	&Event{TypeMeta: TypeMeta{Kind: EventKind}},
	&Player{TypeMeta: TypeMeta{Kind: PlayerKind}},
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

// func DecodePayloadRequest(meta *TypeMeta, payload []byte) (Object, error) {
// 	var result Object
// 	for _, obj := range RegisteredTypes {
// 		if meta.Kind == obj.GetObjectKind() {
// 			result = obj.New()
// 			if err := json.Unmarshal(payload, result); err != nil {
// 				return nil, fmt.Errorf("failed decoding payload %v", err)
// 			}
// 		}
// 	}
// 	return result, nil
// }

// func ToResourceName(obj Object) string {
// 	t := reflect.TypeOf(obj)
// 	t = t.Elem()
// 	return strings.ToLower(t.Name())
// }
