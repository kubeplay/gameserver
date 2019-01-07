# Kube Play

A gamefication platform for learning Kubernetes and Cloud Native apps

This is a working in progress project, still prototype phase, please come back later

# Quick Start

```bash
# Start Game Server
JWT_SECRET=goo go run cmd/server/gameserver.go
# Build kubeplayctl
go build -o /usr/local/bin/kubeplay cmd/kubeplayctl/kubeplayctl.go
# Add an event
kubeplay create -f examples/event.yaml
# Add a challenge
kubeplay create -f examples/challenge.yaml
# Create a new game
kubeplay create game -e meetup --challenge foo
# [HOST] Start a game
# NOTE: A player cannot start a game, it must be started automatically when deploying the game
kubeplay start <event>/<gamename>
# [HOST] Hack the game using pre computed game keys
# NOTE: The game is responsible to inject those keys during the challenge, this is used as a help utility only.
kubeplay hack <event>/<gamename>
# Solve game keys
kubeplay solve <event>/<gamename> <gamekey>
```

