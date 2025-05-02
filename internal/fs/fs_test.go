package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/runn/internal/scope"
)

func TestPath(t *testing.T) {
	currentGlobalCacheDir := GetCacheDir()
	tempDir := t.TempDir()
	SetGlobalCacheDir(tempDir)
	t.Cleanup(func() {
		SetGlobalCacheDir(currentGlobalCacheDir)
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
			filepath.Join(tempDir, "path/to/book.yml"),
			false,
			true,
			"",
			true,
		},
		{
			"allow scope `read:remote`",
			filepath.Join(tempDir, "path/to/book.yml"),
			true,
			false,
			filepath.Join(tempDir, "path/to/book.yml"),
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
			var scopes []string
			if tt.readParent {
				scopes = append(scopes, scope.AllowReadParent)
			} else {
				scopes = append(scopes, scope.DenyReadParent)
			}
			if tt.readRemote {
				scopes = append(scopes, scope.AllowReadRemote)
			} else {
				scopes = append(scopes, scope.DenyReadRemote)
			}
			if err := scope.Set(scopes...); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := scope.Set(scope.DenyReadParent, scope.DenyReadRemote); err != nil {
					t.Fatal(err)
				}
			})

			got, err := Path(tt.p, root)
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
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd)) // Go up two levels from internal/fs to get to the project root

	// Set both read:remote and read:parent scopes
	if err := scope.Set(scope.AllowReadRemote, scope.AllowReadParent); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		pathp   string
		want    int
		wantErr bool
	}{
		{filepath.Join(projectRoot, "testdata/book/book.yml"), 1, false},
		{filepath.Join(projectRoot, "testdata/book/notexist.yml"), 0, true},
		{filepath.Join(projectRoot, "testdata/book/runn_*"), 4, false},
		{filepath.Join(projectRoot, "testdata/book/book.yml") + ":" + filepath.Join(projectRoot, "testdata/book/http.yml"), 2, false},
		{filepath.Join(projectRoot, "testdata/book/book.yml") + ":" + filepath.Join(projectRoot, "testdata/book/runn_*.yml"), 5, false},
		{filepath.Join(projectRoot, "testdata/book/book.yml") + ":" + filepath.Join(projectRoot, "testdata/book/book.yml"), 1, false},
		{filepath.Join(projectRoot, "testdata/book/runn_0_success.yml") + ":" + filepath.Join(projectRoot, "testdata/book/runn_*.yml"), 4, false},
		{"github://k1LoW/runn/testdata/book/book.yml", 1, false},
		{"github://k1LoW/runn/testdata/book/runn_*", 4, false},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/book.yml", 1, false},
		{"file://" + filepath.Join(projectRoot, "testdata/book/book.yml"), 1, false},
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

	t.Cleanup(func() {
		if err := RemoveCacheDir(); err != nil {
			t.Fatal(err)
		}
		if err := scope.Set(scope.DenyReadParent, scope.DenyReadRemote); err != nil {
			t.Fatal(err)
		}
	})
	for _, tt := range tests {
		t.Run(tt.pathp, func(t *testing.T) {
			paths, err := FetchPaths(tt.pathp)
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
