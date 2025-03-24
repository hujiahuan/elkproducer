package elkproducer

import (
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"net/http"
	"time"
)

type ESConfig struct {
	ESConf    Config
	Index     string
	IndexType string
	Url       string
	From      int
	Size      int
	DebugMode bool
}

type Config struct {
	Addresses []string // A list of Elasticsearch nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.

	CloudID                string // Endpoint for the Elastic Service (https://elastic.co/cloud).
	APIKey                 string // Base64-encoded token for authorization; if set, overrides username/password and service token.
	ServiceToken           string // Service token for authorization; if set, overrides username/password.
	CertificateFingerprint string // SHA256 hex fingerprint given by Elasticsearch on first launch.

	Header http.Header // Global HTTP request header.

	// PEM-encoded certificate authorities.
	// When set, an empty certificate pool will be created, and the certificates will be appended to it.
	// The option is only valid when the transport is not specified, or when it's http.Transport.
	CACert []byte

	RetryOnStatus []int                           // List of status codes for retry. Default: 502, 503, 504.
	DisableRetry  bool                            // Default: false.
	MaxRetries    int                             // Default: 3.
	RetryOnError  func(*http.Request, error) bool // Optional function allowing to indicate which error should be retried. Default: nil.

	CompressRequestBody bool // Default: false.
	//CompressRequestBodyLevel int  // Default: gzip.DefaultCompression.

	DiscoverNodesOnStart  bool          // Discover nodes when initializing the client. Default: false.
	DiscoverNodesInterval time.Duration // Discover nodes periodically. Default: disabled.

	EnableMetrics           bool // Enable the metrics collection.
	EnableDebugLogger       bool // Enable the debug logging.
	EnableCompatibilityMode bool // Enable sends compatibility header

	DisableMetaHeader bool // Disable the additional "X-Elastic-Client-Meta" HTTP header.

	RetryBackoff func(attempt int) time.Duration // Optional backoff duration. Default: nil.

	Transport http.RoundTripper         // The HTTP transport object.
	Logger    elastictransport.Logger   // The logger object.
	Selector  elastictransport.Selector // The selector object.

	// Optional constructor function for a custom ConnectionPool. Default: nil.
	ConnectionPoolFunc func([]*elastictransport.Connection, elastictransport.Selector) elastictransport.ConnectionPool
}
