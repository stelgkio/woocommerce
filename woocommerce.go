package woocommerce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/google/go-querystring/query"
)

const (
	UserAgent            = "woocommerce/1.0.0"
	defaultHttpTimeout   = 30
	defaultApiPathPrefix = "/wp-json/wc/v3"
	defaultVersion       = "v3"
)

var (
	apiVersionRegex = regexp.MustCompile(`^v[0-9]{2}$`)
)

type App struct {
	CustomerKey    string
	CustomerSecret string
	AppName        string
	UserId         string
	Scope          string
	ReturnUrl      string
	CallbackUrl    string
	Client         *Client
}

type RateLimitInfo struct {
	RequestCount      int
	BucketSize        int
	RetryAfterSeconds float64
}

type Client struct {
	Client     *http.Client
	app        App
	version    string
	log        LeveledLoggerInterface
	baseURL    *url.URL
	pathPrefix string
	token      string

	// max number of retries, defaults to 0 for no retries see WithRetry option
	retries  int
	attempts int

	RateLimits       RateLimitInfo
	Product          ProductService
	ProductVariation ProductVariationService
	Customer         CustomerService
	Order            OrderService
	OrderNote        OrderNoteService
	Webhook          WebhookService
	PaymentGateway   PaymentGatewayService
	Report           ReportService
}

// NewClient returns a new WooCommerce API client with an already authenticated shopname and
// token. The shopName parameter is the shop's wooCommerce website domain,
// e.g. "shop.gitvim.com"
// a.NewClient(shopName, token, opts) is equivalent to NewClient(a, shopName, token, opts)
func (a App) NewClient(shopName string, opts ...Option) *Client {
	return NewClient(a, shopName, opts...)
}

// NewClient Returns a new WooCommerce API client with an already authenticated shopname and
// token. The shopName parameter is the shop's wooCommerce website domain,
// e.g. "shop.gitvim.com"
func NewClient(app App, shopName string, opts ...Option) *Client {
	baseURL, err := url.Parse(shopName)
	if err != nil {
		panic(err)
	}
	c := &Client{
		Client: &http.Client{
			Timeout: time.Second * defaultHttpTimeout,
		},
		log:        &LeveledLogger{},
		app:        app,
		baseURL:    baseURL,
		version:    defaultVersion,
		pathPrefix: defaultApiPathPrefix,
	}

	c.Product = &ProductServiceOp{client: c}
	c.ProductVariation = &ProductVariationServiceOp{client: c}
	c.Customer = &CustomerServiceOp{client: c}
	c.Order = &OrderServiceOp{client: c}
	c.OrderNote = &OrderNoteServiceOp{client: c}
	c.Webhook = &WebhookServiceOp{client: c}
	c.PaymentGateway = &PaymentGatewayServiceOp{client: c}
	c.Report = &ReportServiceOp{client: c}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ShopBaseURL return a shop's base https base url
func ShopBaseURL(shopName string) string {
	return fmt.Sprintf("https://%s", shopName)
}

// Do sends an API request and populates the given interface with the parsed
// response. It does not make much sense to call Do without a prepared
// interface instance.
func (c *Client) Do(req *http.Request, v interface{}) error {
	_, err := c.doGetHeaders(req, v)
	if err != nil {
		return err
	}

	return nil
}

// doGetHeaders executes a request, decoding the response into `v` and also returns any response headers.
func (c *Client) doGetHeaders(req *http.Request, v interface{}) (http.Header, error) {
	var resp *http.Response
	var err error

	retries := c.retries
	c.attempts = 0
	c.logRequest(req)
	// Check if the scheme is "https"
	if req.URL.Scheme == "https" {
		q := req.URL.Query()
		q.Set("consumer_key", c.app.CustomerKey)
		q.Set("consumer_secret", c.app.CustomerSecret)
		req.URL.RawQuery = q.Encode()
		//fmt.Println("The URL is HTTPS")
	} else {
		// Create a new OAuth1 configuration
		config := oauth1.NewConfig(c.app.CustomerKey, c.app.CustomerSecret)
		token := oauth1.NewToken("", "")

		// Create an OAuth1 HTTP client
		c.Client = config.Client(oauth1.NoContext, token)
		fmt.Println("The URL is not HTTPS", req.URL.Scheme)
	}
	for {
		c.attempts++

		resp, err = c.Client.Do(req)

		c.logResponse(resp)
		if err != nil {
			return nil, err //http client errors, not api responses
		}

		respErr := CheckResponseError(resp)
		if respErr == nil {
			break // no errors, break out of the retry loop
		}

		// retry scenario, close resp and any continue will retry
		resp.Body.Close()

		if retries <= 1 {
			return nil, respErr
		}

		if rateLimitErr, isRetryErr := respErr.(RateLimitError); isRetryErr {
			wait := time.Duration(rateLimitErr.RetryAfter) * time.Second
			c.log.Debugf("rate limited waiting %s", wait.String())
			time.Sleep(wait)
			retries--
			continue
		}

		var doRetry bool
		switch resp.StatusCode {
		case http.StatusServiceUnavailable:
			c.log.Debugf("service unavailable, retrying")
			doRetry = true
			retries--
		}

		if doRetry {
			continue
		}

		//fmt.Println(respErr, "err result", resp)
		// no retry attempts, just return the err
		return nil, respErr
	}

	c.logResponse(resp)
	defer resp.Body.Close()

	if v != nil {
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&v)
		if err != nil {
			return nil, err
		}
	}

	return resp.Header, nil
}

// ResponseDecodingError occurs when the response body from WooCommerce could
// not be parsed.
type ResponseDecodingError struct {
	Body    []byte
	Message string
	Status  int
}

func (e ResponseDecodingError) Error() string {
	return e.Message
}

func CheckResponseError(r *http.Response) error {
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	// Create an anonoymous struct to parse the JSON data into.
	woocommerceError := struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// empty body, this probably means WooCommerce returned an error with no body
	// we'll handle that error in wrapSpecificError()
	if len(bodyBytes) > 0 {
		err := json.Unmarshal(bodyBytes, &woocommerceError)
		if err != nil {
			return ResponseDecodingError{
				Body:    bodyBytes,
				Message: err.Error(),
				Status:  r.StatusCode,
			}
		} else {
			return ResponseError{
				Status:  r.StatusCode,
				Message: woocommerceError.Message,
			}
		}
	}

	// Create the response error from the WooCommerce error.
	responseError := ResponseError{
		Status:  r.StatusCode,
		Message: woocommerceError.Message,
	}

	// If the errors field is not filled out, we can return here.
	if woocommerceError.Message == "" {
		return wrapSpecificError(r, responseError)
	}

	// 	switch reflect.TypeOf(woocommerceError.Errors).Kind() {
	// 	case reflect.String:
	// 		// Single string, use as message
	// 		responseError.Message = woocommerceError.Errors.(string)
	// 	case reflect.Slice:
	// 		// An array, parse each entry as a string and join them on the message
	// 		// json always serializes JSON arrays into []interface{}
	// 		for _, elem := range woocommerceError.Errors.([]interface{}) {
	// 			responseError.Data = append(responseError.Data, fmt.Sprint(elem))
	// 		}
	// 		responseError.Message = strings.Join(responseError.Data, ", ")
	// 	case reflect.Map:
	// 		// A map, parse each error for each key in the map.
	// 		// json always serializes into map[string]interface{} for objects
	// 		for k, v := range woocommerceError.Errors.(map[string]interface{}) {
	// 			// Check to make sure the interface is a slice
	// 			// json always serializes JSON arrays into []interface{}
	// 			if reflect.TypeOf(v).Kind() == reflect.Slice {
	// 				for _, elem := range v.([]interface{}) {
	// 					// If the primary message of the response error is not set, use
	// 					// any message.
	// 					if responseError.Message == "" {
	// 						responseError.Message = fmt.Sprintf("%v: %v", k, elem)
	// 					}
	// 					topicAndElem := fmt.Sprintf("%v: %v", k, elem)
	// 					responseError.Data = append(responseError.Data, topicAndElem)
	// 				}
	// 			}
	// 		}
	// 	}

	return wrapSpecificError(r, responseError)
}

func (c *Client) logRequest(req *http.Request) {
	if req == nil {
		return
	}
	if req.URL != nil {
		c.log.Debugf("%s: %s", req.Method, req.URL.String())
	}
	c.logBody(&req.Body, "SENT: %s")
}

func (c *Client) logResponse(res *http.Response) {
	if res == nil {
		return
	}
	c.log.Debugf("RECV %d: %s", res.StatusCode, res.Status)
	c.logBody(&res.Body, "RESP: %s")
}

func (c *Client) logBody(body *io.ReadCloser, format string) {
	if body == nil {
		return
	}
	b, _ := io.ReadAll(*body)
	if len(b) > 0 {
		c.log.Debugf(format, string(b))
	}
	*body = io.NopCloser(bytes.NewBuffer(b))
}

// ResponseError is A general response error that follows a similar layout to WooCommerce's response
// errors, i.e. either a single message or a list of messages.
// https://woocommerce.github.io/woocommerce-rest-api-docs/#request-response-format
type ResponseError struct {
	Status  int
	Message string
	Data    []string
}

func (e ResponseError) Error() string {
	return e.Message
}

// An error specific to a rate-limiting response. Embeds the ResponseError to
// allow consumers to handle it the same was a normal ResponseError.
type RateLimitError struct {
	ResponseError
	RetryAfter int
}

func wrapSpecificError(r *http.Response, err ResponseError) error {
	if err.Status == http.StatusTooManyRequests {
		f, _ := strconv.ParseFloat(r.Header.Get("Retry-After"), 64)
		return RateLimitError{
			ResponseError: err,
			RetryAfter:    int(f),
		}
	}
	if err.Status == http.StatusNotAcceptable {
		err.Message = http.StatusText(err.Status)
	}

	return err
}

// CreateAndDo performs a web request to WooCommerce with the given method (GET,
// POST, PUT, DELETE) and relative path (e.g. "/wp-admin/v3").
func (c *Client) CreateAndDo(method, relPath string, data, options, resource interface{}) error {
	_, err := c.createAndDoGetHeaders(method, relPath, data, options, resource)
	if err != nil {
		return err
	}
	return nil
}

// createAndDoGetHeaders creates an executes a request while returning the response headers.
func (c *Client) createAndDoGetHeaders(method, relPath string, data, options, resource interface{}) (http.Header, error) {
	if strings.HasPrefix(relPath, "/") {
		relPath = strings.TrimLeft(relPath, "/")
	}

	relPath = path.Join(c.pathPrefix, relPath)
	//println("relPath:", relPath)
	req, err := c.NewRequest(method, relPath, data, options)
	if err != nil {
		return nil, err
	}
	return c.doGetHeaders(req, resource)
}

// Creates an API request. A relative URL can be provided in urlStr, which will
// be resolved to the BaseURL of the Client. Relative URLS should always be
// specified without a preceding slash. If specified, the value pointed to by
// body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, relPath string, body, options interface{}) (*http.Request, error) {
	rel, err := url.Parse(relPath)
	if err != nil {
		return nil, err
	}

	// Make the full url based on the relative path
	u := c.baseURL.ResolveReference(rel)

	// Add custom options
	if options != nil {
		optionsQuery, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		for k, values := range u.Query() {
			for _, v := range values {
				optionsQuery.Add(k, v)
			}
		}
		u.RawQuery = optionsQuery.Encode()
	}

	// A bit of JSON ceremony
	var js []byte = nil

	if body != nil {
		js, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", UserAgent)
	req.SetBasicAuth(c.app.CustomerKey, c.app.CustomerSecret)
	return req, nil
}

// Get performs a GET request for the given path and saves the result in the
// given resource.
func (c *Client) Get(path string, resource, options interface{}) error {
	return c.CreateAndDo("GET", path, nil, options, resource)
}

// Post performs a POST request for the given path and saves the result in the
// given resource.
func (c *Client) Post(path string, data, resource interface{}) error {
	return c.CreateAndDo("POST", path, data, nil, resource)
}

// Put performs a PUT request for the given path and saves the result in the
// given resource.
func (c *Client) Put(path string, data, resource interface{}) error {
	return c.CreateAndDo("PUT", path, data, nil, resource)
}

// Delete performs a DELETE request for the given path
func (c *Client) Delete(path string, options, resource interface{}) error {
	return c.CreateAndDo("DELETE", path, nil, options, resource)
}

// ListOptions represent ist options that can be used for most collections of entities.
type ListOptions struct {
	Context string  `url:"context,omitempty"`
	Page    int     `url:"page,omitempty"`
	PerPage int     `url:"per_page,omitempty"`
	Search  string  `url:"search,omitempty"`
	After   string  `url:"after,omitempty"`
	Before  string  `url:"before,omitempty"`
	Exclude []int64 `url:"exclude,omitempty"`
	Include []int64 `url:"include,omitempty"`
	Offset  int     `url:"offset,omitempty"`
	Order   string  `url:"order,omitempty"`
	Orderby string  `url:"orderby,omitempty"`
}

// DeleteOption is the only option for delete order record. dangerous
// when the force is true, it will permanently delete the resource
// while the force is false, you should get the order from Get Restful API
// but the order's status became to be trash.
// it is better to setting force's column value be "false" rather then  "true"
type DeleteOption struct {
	Force bool `json:"force,omitempty" url:"force,omitempty"`
}

var linkRegex = regexp.MustCompile(`^ *<([^>]+)>; rel="(prev|next|first|last)" *$`)

// Pagination of results
type Pagination struct {
	NextPageOptions     *ListOptions
	PreviousPageOptions *ListOptions
	FirstPageOptions    *ListOptions
	LastPageOptions     *ListOptions
}

// extractPagination extracts pagination info from linkHeader.
// Details on the format are here:
// https://woocommerce.github.io/woocommerce-rest-api-docs/#pagination
// Link: <https://www.example.com/wp-json/wc/v3/products?page=2>; rel="next",
// <https://www.example.com/wp-json/wc/v3/products?page=3>; rel="last"`
func extractPagination(linkHeader string) (*Pagination, error) {
	pagination := new(Pagination)

	if linkHeader == "" {
		return pagination, nil
	}

	for _, link := range strings.Split(linkHeader, ",") {
		match := linkRegex.FindStringSubmatch(link)
		// Make sure the link is not empty or invalid
		println("mm", len(match))
		if len(match) != 4 {
			// We expect 3 values:
			// match[0] = full match
			// match[1] is the URL and match[2] is either 'previous' or 'next', 'first', 'last'
			err := ResponseDecodingError{
				Message: "could not extract pagination link header",
			}
			return nil, err
		}

		rel, err := url.Parse(match[1])
		if err != nil {
			err = ResponseDecodingError{
				Message: "pagination does not contain a valid URL",
			}
			return nil, err
		}

		params, err := url.ParseQuery(rel.RawQuery)
		if err != nil {
			return nil, err
		}

		paginationListOptions := ListOptions{}

		page := params.Get("page")
		if page != "" {
			paginationListOptions.Page, err = strconv.Atoi(params.Get("page"))
			if err != nil {
				return nil, err
			}
		}

		switch match[2] {
		case "next":
			pagination.NextPageOptions = &paginationListOptions
		case "prev":
			pagination.PreviousPageOptions = &paginationListOptions
		case "first":
			pagination.FirstPageOptions = &paginationListOptions
		case "last":
			pagination.LastPageOptions = &paginationListOptions
		}

	}

	return pagination, nil
}
