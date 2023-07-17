package runn

import "testing"

func TestOpenApi3(t *testing.T) {
	c := &httpRunnerConfig{}
	opt := OpenApi3("path/to/openapi3.yml")
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	got := c.OpenApi3DocLocation
	want := "path/to/openapi3.yml"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestOpenApi3FromData(t *testing.T) {
	c := &httpRunnerConfig{}
	opt := OpenApi3FromData([]byte(validOpenApi3Spec))
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	if c.openApi3Doc == nil {
		t.Error("c.openApi3Doc shoud not be nil")
	}
}

func TestSkipValidateRequest(t *testing.T) {
	c := &httpRunnerConfig{}
	opt := SkipValidateRequest(true)
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	got := c.SkipValidateRequest
	want := true
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestSkipValidateResponse(t *testing.T) {
	c := &httpRunnerConfig{}
	opt := SkipValidateResponse(true)
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	got := c.SkipValidateResponse
	want := true
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestMultipartBoundary(t *testing.T) {
	c := &httpRunnerConfig{}
	want := "123456789012345678901234567890abcdefghijklmnopqrstuvwxyz"
	opt := MultipartBoundary(want)
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	got := c.MultipartBoundary
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestHTTPTimeout(t *testing.T) {
	c := &httpRunnerConfig{}
	want := "60s"
	opt := HTTPTimeout(want)
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	got := c.Timeout
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}
