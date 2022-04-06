package runn

import "testing"

func TestOpenApi3(t *testing.T) {
	c := &RunnerConfig{}
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
	c := &RunnerConfig{}
	opt := OpenApi3FromData([]byte(validOpenApi3Spec))
	if err := opt(c); err != nil {
		t.Fatal(err)
	}
	if c.openApi3Doc == nil {
		t.Error("c.openApi3Doc shoud not be nil")
	}
}

func TestSkipValidateRequest(t *testing.T) {
	c := &RunnerConfig{}
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
	c := &RunnerConfig{}
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
