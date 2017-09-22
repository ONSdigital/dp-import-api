package url

// Builder handles building urls in a central place having knowledge of the url structures and base host names.
type Builder struct {
	importAPIHost  string
	datasetAPIHost string
}

// NewBuilder returns a new instance of url.Builder
func NewBuilder(importAPIHost, datasetAPIHost string) *Builder {
	return &Builder{
		importAPIHost:  importAPIHost,
		datasetAPIHost: datasetAPIHost,
	}
}

// GetJobURL returns the url to get a job for the given job ID.
func (builder Builder) GetJobURL(jobID string) string {
	return builder.importAPIHost + "/jobs/" + jobID
}

// GetInstanceURL
func (builder Builder) GetInstanceURL(instanceID string) string {
	return builder.datasetAPIHost + "/instances/" + instanceID
}
