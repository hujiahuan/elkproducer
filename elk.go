package elkproducer

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"time"
)

// elk
func NewClient(cfg ESConfig) (Client, error) {
	es, err := elasticsearch.NewClient(GetElasticConfig(cfg.ESConf))
	if err != nil {
		return nil, err
	}

	if cfg.DebugMode {
		info, err := es.Info()
		if err != nil {
			log.Println("Error getting response: ", err)
		}
		if info.IsError() {
			log.Println("Error: ", info.String())
		} else {
			data, err := json.MarshalIndent(info, "", "\t")
			if err != nil {
				log.Println("Marshal Json Error: ", err)
			} else {
				log.Println(string(data))
			}
		}
	}

	res, err := es.Ping()
	if err != nil {
		log.Println("连接失败: ", err) // 捕获网络级错误（如超时、DNS 解析失败）
		return &ElasticSearchClient{
			es:       es,
			ESConfig: cfg,
		}, err
	} else {
		if res.IsError() {
			log.Println("ES 返回错误: ", res.String()) // 捕获服务级错误（如认证失败）
			return &ElasticSearchClient{
				es:       es,
				ESConfig: cfg,
			}, errors.New(res.String())
		} else {
			return &ElasticSearchClient{
				es:       es,
				ESConfig: cfg,
			}, nil
		}
	}
}

func GetElasticConfig(c Config) elasticsearch.Config {
	elastic := elasticsearch.Config{
		Addresses: c.Addresses, // A list of Elasticsearch nodes to use.
		Username:  c.Username,  // Username for HTTP Basic Authentication.
		Password:  c.Password,  // Password for HTTP Basic Authentication.

		CloudID:                c.CloudID,                // Endpoint for the Elastic Service (https://elastic.co/cloud).
		APIKey:                 c.APIKey,                 // Base64-encoded token for authorization; if set, overrides username/password and service token.
		ServiceToken:           c.ServiceToken,           // Service token for authorization; if set, overrides username/password.
		CertificateFingerprint: c.CertificateFingerprint, // SHA256 hex fingerprint given by Elasticsearch on first launch.

		Header: c.Header, // Global HTTP request header.

		// PEM-encoded certificate authorities.
		// When set, an empty certificate pool will be created, and the certificates will be appended to it.
		// The option is only valid when the transport is not specified, or when it's http.Transport.
		CACert:              c.CACert,
		RetryOnStatus:       c.RetryOnStatus, // List of status codes for retry. Default: 502, 503, 504.
		DisableRetry:        c.DisableRetry,
		MaxRetries:          c.MaxRetries,          // Default: 3.
		RetryOnError:        c.RetryOnError,        // Optional function allowing to indicate which error should be retried. Default: nil.
		CompressRequestBody: c.CompressRequestBody, // Default: false.
		//CompressRequestBodyLevel: c.CompressRequestBodyLevel, // Default: gzip.DefaultCompression.
		DiscoverNodesOnStart:    c.DiscoverNodesOnStart,    // Discover nodes when initializing the client. Default: false.
		DiscoverNodesInterval:   c.DiscoverNodesInterval,   // Discover nodes periodically. Default: disabled.
		EnableMetrics:           c.EnableMetrics,           // Enable the metrics collection.
		EnableDebugLogger:       c.EnableDebugLogger,       // Enable the debug logging.
		EnableCompatibilityMode: c.EnableCompatibilityMode, // Enable sends compatibility header
		DisableMetaHeader:       c.DisableMetaHeader,       // Disable the additional "X-Elastic-Client-Meta" HTTP header.
		RetryBackoff:            c.RetryBackoff,            // Optional backoff duration. Default: nil.
		Transport:               c.Transport,               // The HTTP transport object.
		Logger:                  c.Logger,                  // The logger object.
		Selector:                c.Selector,                // The selector object.
		// Optional constructor function for a custom ConnectionPool. Default: nil.
		ConnectionPoolFunc: c.ConnectionPoolFunc,
	}
	return elastic
}

type ElasticSearchClient struct {
	es       *elasticsearch.Client
	ESConfig ESConfig
}

func (c *ElasticSearchClient) indexName() string {
	//tl := "2006-01-02"
	//indexDate := time.Now().Format(tl)
	//return strings.Trim(strings.Trim(c.ytESConfig.IndexPrefix, "-"), "_") + "-" + indexDate
	sum := md5.Sum([]byte(c.ESConfig.Url))
	return c.ESConfig.Index + hex.EncodeToString(sum[:])
}

func (c *ElasticSearchClient) AddDoc(doc interface{}) {
	body, err := json.Marshal(doc)
	if err != nil {
		log.Println(err)
		return
	}
	req := esapi.IndexRequest{
		Index: c.indexName(),
		Body:  bytes.NewReader(body),
	}
	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		log.Println(err)
		return
	}
	if c.ESConfig.DebugMode {
		if res.IsError() {
			log.Println("Error parsing response body: ", err)
		}
	}
	return
}

func (c *ElasticSearchClient) AddDocAsync(doc interface{}) {
	go c.AddDoc(doc)
}

func (c *ElasticSearchClient) AddLog(logBody interface{}) {
	logDoc := &LogDocument{
		Timestamp: time.Now(),
		Log:       logBody,
		Url:       c.ESConfig.Url,
	}
	body, err := json.Marshal(logDoc)
	if err != nil {
		log.Println(err)
		return
	}
	req := esapi.IndexRequest{
		Index:   c.indexName(),
		Body:    bytes.NewReader(body),
		Refresh: "true",
	}
	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		log.Println(err)
		return
	}
	if c.ESConfig.DebugMode {
		log.Println(string(body))
		if res.IsError() {
			log.Println("Error parsing response body: ", err)
		}
	}
	return
}

func (c *ElasticSearchClient) AddLogAsync(logBody interface{}) {
	go c.AddLog(logBody)
}

func (c *ElasticSearchClient) GetDoc() {

}

func (c *ElasticSearchClient) GetDocAsync() {
	go c.GetDoc()
}

func (c *ElasticSearchClient) GetLog() map[string]interface{} {
	var (
		r map[string]interface{}
	)
	var buf bytes.Buffer
	query := map[string]interface{}{
		"from": c.ESConfig.From,
		"size": c.ESConfig.Size,
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"url": c.ESConfig.Url,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := c.es.Search(
		c.es.Search.WithContext(context.Background()),
		c.es.Search.WithIndex(c.indexName()),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
		c.es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if c.ESConfig.DebugMode {
		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Printf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
	}
	return r
	// Print the response status, number of results, and request duration.
	//log.Printf(
	//	"[%s] %d hits; took: %dms",
	//	res.Status(),
	//	int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
	//	int(r["took"].(float64)),
	//)
	// Print the ID and document source for each hit.
	//for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
	//	log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	//}
	//log.Println(strings.Repeat("=", 37))
}

func (c *ElasticSearchClient) GetLogAsync() {
	go c.GetLog()
}

func (c *ElasticSearchClient) GetData(q map[string]interface{}) map[string]interface{} {
	var (
		r map[string]interface{}
	)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := c.es.Search(
		c.es.Search.WithContext(context.Background()),
		c.es.Search.WithIndex(c.ESConfig.Index),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
		c.es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if c.ESConfig.DebugMode {
		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Printf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
	}
	return r
	// Print the response status, number of results, and request duration.
	//log.Printf(
	//	"[%s] %d hits; took: %dms",
	//	res.Status(),
	//	int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
	//	int(r["took"].(float64)),
	//)
	// Print the ID and document source for each hit.
	//for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
	//	log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	//}
	//log.Println(strings.Repeat("=", 37))
}

func (c *ElasticSearchClient) GetTeeLog() map[string]interface{} {
	var (
		r map[string]interface{}
	)
	var buf bytes.Buffer
	query := map[string]interface{}{
		"from": c.ESConfig.From,
		"size": c.ESConfig.Size,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := c.es.Search(
		c.es.Search.WithContext(context.Background()),
		c.es.Search.WithIndex(c.ESConfig.Index),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
		c.es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if c.ESConfig.DebugMode {
		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Printf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
	}
	return r
	// Print the response status, number of results, and request duration.
	//log.Printf(
	//	"[%s] %d hits; took: %dms",
	//	res.Status(),
	//	int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
	//	int(r["took"].(float64)),
	//)
	// Print the ID and document source for each hit.
	//for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
	//	log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	//}
	//log.Println(strings.Repeat("=", 37))
}

func (c *ElasticSearchClient) Ping() bool {
	res, err := c.es.Ping()
	if err != nil {
		log.Println("连接失败: ", err) // 捕获网络级错误（如超时、DNS 解析失败）
		return false
	} else {
		if res.IsError() {
			log.Println("ES 返回错误: ", res.String()) // 捕获服务级错误（如认证失败）
			return false
		} else {
			return true
		}
	}
}
