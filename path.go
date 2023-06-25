package runn

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/k1LoW/ghfs"
	"github.com/k1LoW/urlfilepath"
)

const (
	prefixHttps  = "https://"
	prefixGitHub = "github://"
)

func ShortenPath(p string) string {
	flags := strings.Split(p, string(filepath.Separator))
	abs := false
	if flags[0] == "" {
		abs = true
	}
	var s []string
	for _, f := range flags[:len(flags)-1] {
		if len(f) > 0 {
			s = append(s, string(f[0]))
		}
	}
	s = append(s, flags[len(flags)-1])
	if abs {
		return string(filepath.Separator) + filepath.Join(s...)
	}
	return filepath.Join(s...)
}

func fetchPaths(pathp string) ([]string, error) {
	paths := []string{}
	listp := splitList(pathp)
	for _, pp := range listp {
		base, pattern := doublestar.SplitPattern(filepath.ToSlash(pp))
		var fsys fs.FS
		fetchRequired := false
		fetchDir := ""
		switch {
		case strings.HasPrefix(base, prefixHttps):
			// https://
			if strings.Contains(pattern, "*") {
				return nil, fmt.Errorf("https scheme does not support wildcard: %s", pp)
			}
			p, err := fetchHTTPSBook(pp)
			if err != nil {
				return nil, err
			}
			paths = append(paths, p)
			continue
		case strings.HasPrefix(base, prefixGitHub):
			// github://
			fetchRequired = true
			splitted := strings.Split(strings.TrimPrefix(base, prefixGitHub), "/")
			if len(splitted) < 2 {
				return nil, fmt.Errorf("invalid path: %s", pp)
			}
			owner := splitted[0]
			repo := splitted[1]
			sub := splitted[2:]
			gfs, err := ghfs.New(owner, repo)
			if err != nil {
				return nil, err
			}
			if len(sub) > 0 {
				fsys, err = gfs.Sub(strings.Join(sub, "/"))
				if err != nil {
					return nil, err
				}
			} else {
				fsys = gfs
			}
			cd, err := cacheDir()
			if err != nil {
				return nil, err
			}
			u, err := url.Parse(base)
			if err != nil {
				return nil, err
			}
			ep, err := urlfilepath.Encode(u)
			if err != nil {
				return nil, err
			}
			fetchDir = filepath.Join(cd, ep)
		default:
			// local file system
			abs, err := filepath.Abs(base)
			if err != nil {
				return nil, err
			}
			fsys = os.DirFS(abs)
		}
		if err := doublestar.GlobWalk(fsys, pattern, func(p string, d fs.DirEntry) error {
			if d.IsDir() {
				return nil
			}
			if !fetchRequired {
				paths = append(paths, filepath.Join(base, p))
				return nil
			}
			f, err := fsys.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			p = filepath.Join(fetchDir, p)
			if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
				return err
			}
			n, err := os.Create(p)
			if err != nil {
				return err
			}
			defer n.Close()
			if _, err = io.Copy(n, f); err != nil {
				return err
			}
			paths = append(paths, p)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return unique(paths), nil
}

func splitList(pathp string) []string {
	rep := strings.NewReplacer(prefixHttps, repKey(prefixHttps), prefixGitHub, repKey(prefixGitHub))
	per := strings.NewReplacer(repKey(prefixHttps), prefixHttps, repKey(prefixGitHub), prefixGitHub)
	listp := []string{}
	for _, p := range filepath.SplitList(rep.Replace(pathp)) {
		listp = append(listp, per.Replace(p))
	}
	return listp
}

func fetchHTTPSBook(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	cd, err := cacheDir()
	if err != nil {
		return "", err
	}
	ep, err := urlfilepath.Encode(u)
	if err != nil {
		return "", err
	}
	p := filepath.Join(cd, ep)
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return "", err
	}
	n, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer n.Close()
	if _, err = io.Copy(n, res.Body); err != nil {
		return "", err
	}
	return p, nil
}

func readFile(name string) ([]byte, error) {
	if globalCacheDir == "" || !strings.HasPrefix(name, globalCacheDir) {
		return os.ReadFile(name)
	}
	if _, err := os.Stat(name); err == nil {
		return os.ReadFile(name)
	}
	pathstr, err := filepath.Rel(globalCacheDir, name)
	if err != nil {
		return nil, err
	}
	u, err := urlfilepath.Decode(pathstr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "https":
		b, err := readFileViaHTTPS(u.String())
		if err != nil {
			return nil, err
		}
		// write cache
		if err := os.WriteFile(name, b, os.ModePerm); err != nil {
			return nil, err
		}
		return b, err
	case "github":
		b, err := readFileViaGitHub(u.String())
		if err != nil {
			return nil, err
		}
		// write cache
		if err := os.WriteFile(name, b, os.ModePerm); err != nil {
			return nil, err
		}
		return b, err
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.String())
	}
}

func fetchPath(path string) (string, error) {
	paths, err := fetchPaths(path)
	if err != nil {
		return "", err
	}
	if len(paths) != 1 {
		return "", errors.New("invalid path")
	}
	return paths[0], nil
}

func readFileViaHTTPS(urlstr string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, urlstr, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func readFileViaGitHub(urlstr string) ([]byte, error) {
	splitted := strings.Split(strings.TrimPrefix(urlstr, prefixGitHub), "/")
	if len(splitted) < 2 {
		return nil, fmt.Errorf("invalid url: %s", urlstr)
	}
	owner := splitted[0]
	repo := splitted[1]
	p := strings.Join(splitted[2:], "/")
	gfs, err := ghfs.New(owner, repo)
	if err != nil {
		return nil, err
	}
	f, err := gfs.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func unique(in []string) []string {
	u := []string{}
	m := map[string]struct{}{}
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		u = append(u, s)
		m[s] = struct{}{}
	}
	return u
}

func repKey(in string) string {
	return fmt.Sprintf("RUNN_%s_SCHEME", strings.TrimSuffix(strings.ToUpper(in), "://"))
}
