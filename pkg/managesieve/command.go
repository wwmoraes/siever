package managesieve

type Command[T any] interface {
	Execute(*baseClient, ...any) (T, string, error)
}

type RawCommand string

func (cmd *RawCommand) Execute(client *baseClient, args ...any) ([]string, string, error) {
	return client.RawCmd(string(*cmd), args...)
}
