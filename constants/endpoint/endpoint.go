package endpoint

import "strings"

type Format int64

const (
	NoSlash Format = iota
	LeadingSlash
	TrailingSlash
	LeadingAndTrailingSlash
)

type Endpoint struct {
	path string
}

func New(path string) Endpoint {
	e := Endpoint{path: strings.Trim(strings.Trim(path, " "), "/")}
	return e
}

func Join(base Endpoint, path string) Endpoint {
	e := Endpoint{path: base.FormattedPath(TrailingSlash) + strings.Trim(strings.Trim(path, " "), "/")}
	return e
}

func (m Endpoint) Path() string {
	return m.FormattedPath(LeadingSlash)
}

func (m Endpoint) FormattedPath(format Format) string {
	var path = strings.Trim(m.path, " ")
	switch format {
	case NoSlash:
		return path
	case LeadingSlash:
		return "/" + path
	case TrailingSlash:
		return path + "/"
	case LeadingAndTrailingSlash:
		return "/" + path + "/"
	}

	return path
}
