package elkproducer

type Client interface {
	// AddDocAsync add generic doc to elk
	AddDocAsync(doc interface{})
	// AddLogAsync add whole log body as "log" field with auto timestamp to elk
	AddLogAsync(log interface{})
	//simple thead log
	AddLog(log interface{})
	//simple thead doc
	AddDoc(doc interface{})
	//get doc from elk
	GetDocAsync()
	//get log from elk
	GetLogAsync()
	//simple thead get doc
	GetDoc()
	//simple thead get log
	GetLog() map[string]interface{}

	GetTeeLog() map[string]interface{}
	//query
	GetData(map[string]interface{}) map[string]interface{}
}
