package runn

import (
	"errors"
	"strings"
	"sync"
)

type scopes struct {
	readParent bool
	readRemote bool
	runExec    bool
	mu         sync.RWMutex
}

const (
	ScopeAllowReadParent = "read:parent"
	ScopeAllowReadRemote = "read:remote"
	ScopeAllowRunExec    = "run:exec" //nostyle:repetition
	ScopeDenyReadParent  = "!read:parent"
	ScopeDenyReadRemote  = "!read:remote"
	ScopeDenyRunExec     = "!run:exec" //nostyle:repetition
)

var ErrInvalidScope = errors.New("invalid scope")

var globalScopes = &scopes{
	readParent: false,
	readRemote: false,
	runExec:    false,
}

func setScopes(scopes ...string) error {
	globalScopes.mu.Lock()
	defer globalScopes.mu.Unlock()
	for _, s := range scopes {
		splitted := strings.Split(strings.TrimSpace(s), ",")
		for _, ss := range splitted {
			switch ss {
			case ScopeAllowReadParent:
				globalScopes.readParent = true
			case ScopeAllowReadRemote:
				globalScopes.readRemote = true
			case ScopeAllowRunExec:
				globalScopes.runExec = true
			case ScopeDenyReadParent:
				globalScopes.readParent = false
			case ScopeDenyReadRemote:
				globalScopes.readRemote = false
			case ScopeDenyRunExec:
				globalScopes.runExec = false
			case "":
			default:
				return ErrInvalidScope
			}
		}
	}
	return nil
}
