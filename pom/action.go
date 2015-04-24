package pom

// Command The gsmake action command interface
type Command interface {
	ExecCommand(args map[string]string) error // Exec command
}

// Action task action object
type Action struct {
	Name    string            // action name
	Args    map[string]string // command input args
	command Command           // the bind command object
}
