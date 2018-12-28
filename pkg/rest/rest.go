package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Interface interface {
	Get() *Request
	Post() *Request
	Put() *Request
	Patch() *Request
	Delete() *Request
	Verb(verb string) *Request
	Reset() *Request
}

// HTTPClient is an interface for testing a request object.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type BasicAuth struct {
	Username string
	Password string
}

// Result contains the result of calling Request.Do().
type Result struct {
	body        []byte
	contentType string
	err         error
	statusCode  int
}

func (r Result) Raw() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if !r.IsSuccess() {
		return nil, fmt.Errorf("failed (%d) performing request to the remote server: %v", r.StatusCode(), string(r.body))
	}
	return r.body, nil
}

func (r Result) Into(obj interface{}) error {
	if r.err != nil {
		return r.err
	}
	if !r.IsSuccess() {
		return fmt.Errorf("failed (%d) performing request to the remote server: %v", r.StatusCode(), string(r.body))
	}
	if err := json.Unmarshal(r.body, obj); err != nil {
		return fmt.Errorf("failed decoding response [%v]", err)
	}
	return nil
}

func (r Result) StatusCode() int {
	return r.statusCode
}

func (r Result) IsSuccess() bool {
	return r.statusCode == 200 || r.statusCode == 201 || r.statusCode == 204 || r.statusCode == 202
}

func (r Result) ContentType() string {
	return r.contentType
}

func (r Result) Error() error {
	return r.err
}

type Request struct {
	Client HTTPClient

	baseURL   *url.URL
	basicAuth *BasicAuth
	verb      string
	timeout   time.Duration
	query     url.Values

	// This is only used for per-request timeouts, deadlines, and cancellations.
	ctx context.Context

	body    io.Reader
	headers http.Header
	err     error
}

func NewRequest(client HTTPClient, baseURL *url.URL) *Request {
	if client == nil {
		client = http.DefaultClient
	}
	request := &Request{
		baseURL: baseURL,
		Client:  client,
		headers: http.Header{"Content-Type": []string{"application/json"}},
		query:   make(url.Values),
	}
	return request
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Set(key, value)
	return r
}

// Timeout makes the request use the given duration as a timeout. Sets the "timeout"
// parameter.
func (r *Request) Timeout(d time.Duration) *Request {
	if r.err != nil {
		return r
	}
	r.timeout = d
	return r
}

// RequestPath override the basePath of the request
func (r *Request) RequestURI(uri ...string) *Request {
	r.baseURL.Path = strings.Join(uri, "/")
	return r
}

func (r *Request) BasicAuth(basicAuth *BasicAuth) *Request {
	r.basicAuth = basicAuth
	return r
}

func (r *Request) Bearer(token interface{}) *Request {
	r.SetHeader("Authorization", fmt.Sprintf("Bearer %v", token))
	return r
}

func (r *Request) AddQuery(key, value string) *Request {
	r.query.Add(key, value)
	return r
}

// Context adds a context to the request. Contexts are only used for
// timeouts, deadlines, and cancellations.
func (r *Request) Context(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) Patch() *Request {
	r.verb = "PATCH"
	return r
}

func (r *Request) Get() *Request {
	r.verb = "GET"
	return r
}

func (r *Request) Post() *Request {
	r.verb = "POST"
	return r
}

func (r *Request) Put() *Request {
	r.verb = "PUT"
	return r
}

func (r *Request) Delete() *Request {
	r.verb = "DELETE"
	return r
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) Body(bodyData interface{}) *Request {
	reqBody, err := json.Marshal(bodyData)
	if err != nil {
		r.err = fmt.Errorf("failed encoding body [%v]", err)
	}
	r.body = bytes.NewBuffer(reqBody)
	return r
}

func (r *Request) Error() error {
	return r.err
}

func (r *Request) Do() *Result {
	result := &Result{}
	if r.err != nil {
		result.err = r.err
		return result
	}
	client := r.Client
	if r.Client == nil {
		client = http.DefaultClient
	}
	// if glog.V(4) {
	// 	glog.Infof("Verb %#v, URL: %#v, URLPath %#v", r.verb, r.URL().String(), r.URL().Path)
	// }

	request, err := http.NewRequest(r.verb, r.baseURL.String(), r.body)
	if err != nil {
		r.err = fmt.Errorf("failed creating request [%v]", err)
		return result
	}
	q := request.URL.Query()
	q = r.query
	request.URL.RawQuery = q.Encode()
	request.Header = r.headers
	if r.basicAuth != nil {
		request.SetBasicAuth(
			r.basicAuth.Username,
			r.basicAuth.Password,
		)
	}
	resp, err := client.Do(request)
	if err != nil {
		result.err = fmt.Errorf("failed processing the request [%v]", err)
		return result
	}
	defer resp.Body.Close()

	result.statusCode = resp.StatusCode
	result.contentType = resp.Header.Get("Content-Type")
	if resp.Body != nil {
		data, err := ioutil.ReadAll(resp.Body)
		// if glog.V(8) {
		// 	glog.Infof("Response Body[%d]: %s", resp.StatusCode, string(data))
		// }
		if err != nil {
			result.err = fmt.Errorf("failed reading response [%v]", err)

			return result
		}
		result.body = data
	}
	return result
}
