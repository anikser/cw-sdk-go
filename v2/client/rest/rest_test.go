package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"code.cryptowat.ch/cw-sdk-go/version"
)

type testCaseREST struct {
	// descr is just a human-readable description of the test case, will be
	// included in the error reports in case of errors
	descr string

	// do is basically a wrapper for any method of RESTClient. As of now it has
	// to perform a single call to the server (actually we might need to remove
	// that limitation if we ever going to compose results from multiple
	// responses), and it should return what RESTClient's method itself has
	// returned.
	do func(c *RESTClient) (interface{}, error)
	// resp is the data which server will respond with, during the do() call.
	resp string

	// wantError is the error we expect to be returned from do(). If it's nil,
	// then wantResult is checked.
	wantError error
	// wantResult is the value we expect to be returned from do().
	wantResult interface{}
}

type testHarnessREST struct {
	t *testing.T
	// checkURL is up to the harness' client; it can check the URL being accessed
	// and complain if it doesn't look right
	checkURL checkURLFunc

	// ts is the test server used to perform HTTP requests
	ts *httptest.Server
	// restClient is an instance of REST client being tested, it has API URL set
	// to the test server's URL
	restClient *RESTClient

	curTestCaseNum int
	curTestCase    *testCaseREST
}

type checkURLFunc func(u *url.URL) error

func newTestHarnessREST(t *testing.T, checkURL checkURLFunc) *testHarnessREST {
	h := &testHarnessREST{
		t:        t,
		checkURL: checkURL,
	}

	h.ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := fmt.Sprintf("cw-sdk-go@%s", version.Version)
		assert.Equal(h.t, userAgent, r.UserAgent())

		if err := h.checkURL(r.URL); err != nil {
			assert.Fail(
				h.t,
				fmt.Sprintf("checkURL returned error: %s", err.Error()),
				"test case #%d (%s)", h.curTestCaseNum, h.curTestCase.descr,
			)
		}
		w.Write([]byte(h.curTestCase.resp))
	}))

	h.restClient = NewRESTClient(&RESTClientParams{
		URL: h.ts.URL,
	})

	return h
}

func (h *testHarnessREST) runTestCases(testCases []testCaseREST) {
	for i, tc := range testCases {
		h.curTestCaseNum = i
		h.curTestCase = &tc
		gotResult, gotError := tc.do(h.restClient)

		assertArgs := []interface{}{
			"test case #%d (%s)", i, tc.descr,
		}

		if tc.wantError != nil {
			if assert.Error(h.t, gotError, assertArgs...) {
				assert.Equal(h.t, tc.wantError, gotError, assertArgs...)
			}
		} else {
			if assert.NoError(h.t, gotError, assertArgs...) {
				assert.Equal(h.t, tc.wantResult, gotResult, assertArgs...)
			}
		}
	}
}

func (h *testHarnessREST) close() {
	h.ts.Close()
}

func dfs(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

// getCheckURL is a helper which returns a checkURLFunc (to pass to
// newTestHarnessREST) which just checks that a URL's path matches wantPath
// exactly.
func getCheckURL(wantPath string) checkURLFunc {
	return func(u *url.URL) error {
		if u.Path != wantPath {
			return errors.Errorf("URL path: want %q, got %q", wantPath, u.Path)
		}

		return nil
	}
}
