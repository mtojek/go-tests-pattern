package go_tests_pattern

import (
	"fmt"
	"io"
	"net/http"
	. "net/http/httptest"
	"testing"
)

/*
   Run with:
   $ go test
*/
func TestRecorder(t *testing.T) {
	type checkFunc func(*ResponseRecorder) error
	check := func(fns ...checkFunc) []checkFunc { return fns }

	hasStatus := func(want int) checkFunc {
		return func(rec *ResponseRecorder) error {
			if rec.Code != want {
				return fmt.Errorf("expected status %d, found %d", want, rec.Code)
			}
			return nil
		}
	}
	hasContents := func(want string) checkFunc {
		return func(rec *ResponseRecorder) error {
			if have := rec.Body.String(); have != want {
				return fmt.Errorf("expected body %q, found %q", want, have)
			}
			return nil
		}
	}
	hasHeader := func(key, want string) checkFunc {
		return func(rec *ResponseRecorder) error {
			if have := rec.Result().Header.Get(key); have != want {
				return fmt.Errorf("expected header %s: %q, found %q", key, want, have)
			}
			return nil
		}
	}

	tests := [...]struct {
		name   string
		h      func(w http.ResponseWriter, r *http.Request)
		checks []checkFunc
	}{
		{
			"200 default",
			func(w http.ResponseWriter, r *http.Request) {},
			check(hasStatus(200), hasContents("")),
		},
		{
			"first code only",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.WriteHeader(202)
				w.Write([]byte("hi"))
			},
			check(hasStatus(201), hasContents("hi")),
		},
		{
			"write string",
			func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "hi first")
			},
			check(
				hasStatus(200),
				hasContents("hi first"),
				hasHeader("Content-Type", "text/plain; charset=utf-8"),
			),
		},
	}

	r, _ := http.NewRequest("GET", "http://foo.com/", nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.HandlerFunc(tt.h)
			rec := NewRecorder()
			h.ServeHTTP(rec, r)
			for _, check := range tt.checks {
				if err := check(rec); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
