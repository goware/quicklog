package quicklog

func Noop() Quicklog {
	return &noop{}
}

type noop struct{}

func (n *noop) Info(group string, message string, v ...any) {}
func (n *noop) Warn(group string, message string, v ...any) {}
func (n *noop) Groups() []string                            { return nil }
func (n *noop) Logs(group string) []LogEntry                { return nil }
func (n *noop) ToMap(timezone string, withExactTime bool, group ...string) map[string][]string {
	return map[string][]string{}
}
