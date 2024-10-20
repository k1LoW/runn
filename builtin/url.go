package builtin

import (
	"net/url"
)

func Url(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}
