package pom

// Task project task object
type Task struct {
	Name    string    // task name
	Actions []*Action // task action list
	project *Project  // project which task belongs to
}
