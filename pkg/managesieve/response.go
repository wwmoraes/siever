package managesieve

type Response interface {
	Parse([]string) error
}
