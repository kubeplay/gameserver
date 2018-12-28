package types

import (
	jwt "github.com/dgrijalva/jwt-go"
)

const GameKeyHeaderName = "X-Game-Key"

// /v1/challenges
type Challenge struct {
	TypeMeta `json:",inline"`
	Metadata `json:"metadata"`

	Keys      map[string]Key `json:"keys"`
	AssetsURL string         `json:"assetsURL"`
}

type ChallengeList struct {
	TypeMeta `json:",inline"`
	ListMeta

	Items []Challenge `json:"items"`
}

// /v1/events/<name>
type Event struct {
	TypeMeta `json:",inline"`
	Metadata `json:"metadata"`

	// Paused blocks new games from starting
	Paused bool `json:"paused"`

	// Score *Score `json:"score"`
	// Raking
}

type EventList struct {
	TypeMeta `json:",inline"`
	ListMeta

	Items []Event `json:"items"`
}

// /v1/events/<name>/game
type Game struct {
	TypeMeta `json:",inline"`
	Metadata `json:"metadata"`

	Challenge string     `json:"challenge"`
	Player    string     `json:"player"`
	Status    GameStatus `json:"status,omitempty"`
}

type GameList struct {
	TypeMeta `json:",inline"`
	ListMeta

	Items []Game `json:"items"`
}

type GamePhase string

const (
	GamePending   GamePhase = "Pending"
	GameRunning   GamePhase = "Running"
	GameCompleted GamePhase = "Completed"
)

type GameStatus struct {
	StartTime      string          `json:"startTime"`
	EndTime        string          `json:"endTime"`
	LastSolvedKey  GameKeyStatus   `json:"lastSolvedKey"`
	RegisteredKeys int             `json:"registeredKeys"`
	Phase          GamePhase       `json:"phase"`
	Keys           []GameKeyStatus `json:"keys"`
}

type GameKeyStatus struct {
	KeyName    string  `json:"keyName"`
	Approved   bool    `json:"approved"`
	ApprovedAt string  `json:"approvedAt,omitempty"`
	Weight     float32 `json:"weight"`
}

type Key struct {
	Value       string  `json:"value,omitempty"`
	Description string  `json:"description"`
	Weight      float32 `json:"weight"`
}

// /v1/player
type Player struct {
	TypeMeta `json:",inline"`
	Metadata `json:"metadata"`

	FirstName string `json:"firstName"`
	SurName   string `json:"surName"`
	Email     string `json:"email"`
}

type PlayerList struct {
	TypeMeta `json:",inline"`
	ListMeta

	Items []Player `json:"items"`
}

type PlayerClaims struct {
	Name      string `json:"name"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Location  string `json:"location"`
	Email     string `json:"email"`

	AccessToken string `json:"access_token,omitempty"`

	jwt.StandardClaims `json:"-"`
}
