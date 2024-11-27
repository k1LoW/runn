package runn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFp(t *testing.T) {
	currentGlobalCacheDir := globalCacheDir
	globalCacheDir = t.TempDir()
	t.Cleanup(func() {
		globalCacheDir = currentGlobalCacheDir
	})
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		p          string
		readRemote bool
		readParent bool
		want       string
		wantErr    bool
	}{
		{
			"Join root and path",
			"path/to/book.yml",
			false,
			false,
			filepath.Join(root, "path/to/book.yml"),
			false,
		},
		{
			"scope `read:parent` error",
			"/path/to/book.yml",
			false,
			false,
			"",
			true,
		},
		{
			"allow scope `read:parent`",
			"/path/to/book.yml",
			false,
			true,
			"/path/to/book.yml",
			false,
		},
		{
			"Join root and path with relative path",
			"path/../book.yml",
			false,
			false,
			filepath.Join(root, "book.yml"),
			false,
		},
		{
			"scope `read:parent` error with relative path",
			"../book.yml",
			false,
			false,
			"",
			true,
		},
		{
			"allow scope `read:parent` with relative path",
			"../book.yml",
			false,
			true,
			filepath.Join(filepath.Dir(root), "book.yml"),
			false,
		},
		{
			"scope `read:remote` error",
			filepath.Join(globalCacheDir, "path/to/book.yml"),
			false,
			true,
			"",
			true,
		},
		{
			"allow scope `read:remote`",
			filepath.Join(globalCacheDir, "path/to/book.yml"),
			true,
			false,
			filepath.Join(globalCacheDir, "path/to/book.yml"),
			false,
		},
		{
			"Join root and path with file://",
			"file://path/to/book.yml",
			false,
			false,
			filepath.Join(root, "path/to/book.yml"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalScopes.mu.Lock()
			globalScopes.readParent = tt.readParent
			globalScopes.readRemote = tt.readRemote
			globalScopes.mu.Unlock()
			t.Cleanup(func() {
				globalScopes.mu.Lock()
				globalScopes.readParent = false
				globalScopes.readRemote = false
				globalScopes.mu.Unlock()
			})

			got, err := fp(tt.p, root)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("want error")
				return
			}

			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}

}

func TestFetchPaths(t *testing.T) {
	tests := []struct {
		pathp   string
		want    int
		wantErr bool
	}{
		{"testdata/book/book.yml", 1, false},
		{"testdata/book/notexist.yml", 0, true},
		{"testdata/book/runn_*", 4, false},
		{"testdata/book/book.yml:testdata/book/http.yml", 2, false},
		{"testdata/book/book.yml:testdata/book/runn_*.yml", 5, false},
		{"testdata/book/book.yml:testdata/book/book.yml", 1, false},
		{"testdata/book/runn_0_success.yml:testdata/book/runn_*.yml", 4, false},
		{"github://k1LoW/runn/testdata/book/book.yml", 1, false},
		{"github://k1LoW/runn/testdata/book/runn_*", 4, false},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/book.yml", 1, false},
		{"file://testdata/book/book.yml", 1, false},
	}

	if os.Getenv("CI") == "" {
		// GITHUB_TOKEN for GitHub Actions does not have permission to access Gist
		tests = append(tests, []struct {
			pathp   string
			want    int
			wantErr bool
		}{
			// Single file
			{"gist://b908ae0721300ca45f4e8b81b6be246d", 1, false},
			{"gist://b908ae0721300ca45f4e8b81b6be246d/book.yml", 1, false},
			// Multiple files
			{"gist://def6fa739fba3fcf211b018f41630adc", 0, true},
			{"gist://def6fa739fba3fcf211b018f41630adc/book.yml", 1, false},
		}...)
	}
	globalScopes.mu.RLock()
	globalScopes.readRemote = true
	globalScopes.mu.RUnlock()

	t.Cleanup(func() {
		if err := RemoveCacheDir(); err != nil {
			t.Fatal(err)
		}
		globalScopes.mu.RLock()
		globalScopes.readRemote = false
		globalScopes.mu.RUnlock()
	})
	for _, tt := range tests {
		t.Run(tt.pathp, func(t *testing.T) {
			paths, err := fetchPaths(tt.pathp)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("want err")
			}
			got := len(paths)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"path/to/book.yml", "p/t/book.yml"},
		{"book.yml", "book.yml"},
		{"/path/to/book.yml", "/p/t/book.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ShortenPath(tt.in)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
