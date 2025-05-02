package scope

import (
	"testing"
)

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		scopes  []string
		wantErr bool
	}{
		{
			name:    "set allow read parent",
			scopes:  []string{AllowReadParent},
			wantErr: false,
		},
		{
			name:    "set allow read remote",
			scopes:  []string{AllowReadRemote},
			wantErr: false,
		},
		{
			name:    "set allow run exec",
			scopes:  []string{AllowRunExec},
			wantErr: false,
		},
		{
			name:    "set deny read parent",
			scopes:  []string{DenyReadParent},
			wantErr: false,
		},
		{
			name:    "set deny read remote",
			scopes:  []string{DenyReadRemote},
			wantErr: false,
		},
		{
			name:    "set deny run exec",
			scopes:  []string{DenyRunExec},
			wantErr: false,
		},
		{
			name:    "set multiple scopes",
			scopes:  []string{AllowReadParent, AllowReadRemote, AllowRunExec},
			wantErr: false,
		},
		{
			name:    "set invalid scope",
			scopes:  []string{"invalid:scope"},
			wantErr: true,
		},
		{
			name:    "set empty scope",
			scopes:  []string{""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset scopes
			Global.readParent = false
			Global.readRemote = false
			Global.runExec = false

			err := Set(tt.scopes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for _, s := range tt.scopes {
					switch s {
					case AllowReadParent:
						if !IsReadParentAllowed() {
							t.Errorf("IsReadParentAllowed() = false, want true")
						}
					case AllowReadRemote:
						if !IsReadRemoteAllowed() {
							t.Errorf("IsReadRemoteAllowed() = false, want true")
						}
					case AllowRunExec:
						if !IsRunExecAllowed() {
							t.Errorf("IsRunExecAllowed() = false, want true")
						}
					case DenyReadParent:
						if IsReadParentAllowed() {
							t.Errorf("IsReadParentAllowed() = true, want false")
						}
					case DenyReadRemote:
						if IsReadRemoteAllowed() {
							t.Errorf("IsReadRemoteAllowed() = true, want false")
						}
					case DenyRunExec:
						if IsRunExecAllowed() {
							t.Errorf("IsRunExecAllowed() = true, want false")
						}
					}
				}
			}
		})
	}
}

func TestScopeFunctions(t *testing.T) {
	// Test IsReadParentAllowed
	Global.readParent = true
	if !IsReadParentAllowed() {
		t.Errorf("IsReadParentAllowed() = false, want true")
	}
	Global.readParent = false
	if IsReadParentAllowed() {
		t.Errorf("IsReadParentAllowed() = true, want false")
	}

	// Test IsReadRemoteAllowed
	Global.readRemote = true
	if !IsReadRemoteAllowed() {
		t.Errorf("IsReadRemoteAllowed() = false, want true")
	}
	Global.readRemote = false
	if IsReadRemoteAllowed() {
		t.Errorf("IsReadRemoteAllowed() = true, want false")
	}

	// Test IsRunExecAllowed
	Global.runExec = true
	if !IsRunExecAllowed() {
		t.Errorf("IsRunExecAllowed() = false, want true")
	}
	Global.runExec = false
	if IsRunExecAllowed() {
		t.Errorf("IsRunExecAllowed() = true, want false")
	}
}
