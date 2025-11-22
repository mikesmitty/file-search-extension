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

// GetModelList returns the list of available models for completion
func GetModelList() []string {
	return []string{
		"gemini-2.5-flash",
		"gemini-2.5-flash-lite",
		"gemini-2.5-pro",
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
	}
}
