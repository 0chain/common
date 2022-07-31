package endpoint

import "strings"

const rootPath = "/"

var RootEndpoint = New(rootPath)

type Format int64

const (
	LeadingSlash Format = iota
	TrailingSlash
	LeadingAndTrailingSlash
	NoSlash
)

type Endpoint struct {
	path string
}

func New(path string) Endpoint {
	var e Endpoint
	var trimmedPath = strings.Trim(path, " ")

	if trimmedPath == rootPath {
		e = Endpoint{path: trimmedPath}
	} else {
		e = Endpoint{path: strings.Trim(trimmedPath, "/")}
	}

	return e
}

func Join(base Endpoint, path string) Endpoint {
	e := New(base.FormattedPath(TrailingSlash) + strings.Trim(strings.Trim(path, " "), "/"))
	return e
}

func (m Endpoint) Path() string {
	return m.FormattedPath(LeadingSlash)
}

func (m Endpoint) FormattedPath(format Format) string {
	var trimmedPath = strings.Trim(m.path, " ")

	if trimmedPath == rootPath {
		return trimmedPath
	}

	switch format {
	case LeadingSlash:
		return "/" + trimmedPath
	case TrailingSlash:
		return trimmedPath + "/"
	case LeadingAndTrailingSlash:
		return "/" + trimmedPath + "/"
	case NoSlash:
		return strings.Trim(trimmedPath, "/")
	}

	return m.Path()
}
