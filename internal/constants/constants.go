package constants

const (
	// DefaultModel is the default Gemini model used for queries
	DefaultModel = "gemini-2.5-flash"

	// Resource Prefixes
	StoreResourcePrefix     = "fileSearchStores/"
	FileResourcePrefix      = "files/"
	DocumentResourcePrefix  = "/documents/"
	OperationResourcePrefix = "/operations/"
)

// GetModelList returns the list of models known to support file search
func GetModelList() []string {
	// Currently only supports these models
	// https://ai.google.dev/gemini-api/docs/file-search#supported-models
	return []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
	}
}
