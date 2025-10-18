package scope

import (
	"errors"
	"strings"
	"sync"
)

// Scopes holds the permission scopes for runn.
type Scopes struct {
	readParent bool
	readRemote bool
	runExec    bool
	mu         sync.RWMutex
}

const (
	// AllowReadParent allows reading files from parent directories.
	AllowReadParent = "read:parent"
	// AllowReadRemote allows reading files from remote sources.
	AllowReadRemote = "read:remote"
	// AllowRunExec allows executing commands.
	AllowRunExec = "run:exec" //nostyle:repetition
	// DenyReadParent denies reading files from parent directories.
	DenyReadParent = "!read:parent"
	// DenyReadRemote denies reading files from remote sources.
	DenyReadRemote = "!read:remote"
	// DenyRunExec denies executing commands.
	DenyRunExec = "!run:exec" //nostyle:repetition
)

// ErrInvalidScope is returned when an invalid scope is provided.
var ErrInvalidScope = errors.New("invalid scope") //nostyle:repetition

// Global is the global scopes instance.
var Global = &Scopes{
	readParent: false,
	readRemote: false,
	runExec:    false,
}

// Set sets the scopes for runn.
func Set(scopes ...string) error {
	Global.mu.Lock()
	defer Global.mu.Unlock()
	for _, s := range scopes {
		splitted := strings.SplitSeq(strings.TrimSpace(s), ",")
		for ss := range splitted {
			switch ss {
			case AllowReadParent:
				Global.readParent = true
			case AllowReadRemote:
				Global.readRemote = true
			case AllowRunExec:
				Global.runExec = true
			case DenyReadParent:
				Global.readParent = false
			case DenyReadRemote:
				Global.readRemote = false
			case DenyRunExec:
				Global.runExec = false
			case "":
			default:
				return ErrInvalidScope
			}
		}
	}
	return nil
}

// IsReadParentAllowed returns whether reading files from parent directories is allowed.
func IsReadParentAllowed() bool {
	Global.mu.RLock()
	defer Global.mu.RUnlock()
	return Global.readParent
}

// IsReadRemoteAllowed returns whether reading files from remote sources is allowed.
func IsReadRemoteAllowed() bool {
	Global.mu.RLock()
	defer Global.mu.RUnlock()
	return Global.readRemote
}

// IsRunExecAllowed returns whether executing commands is allowed.
func IsRunExecAllowed() bool {
	Global.mu.RLock()
	defer Global.mu.RUnlock()
	return Global.runExec
}
