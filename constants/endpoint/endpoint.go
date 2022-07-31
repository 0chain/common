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
	path         string
	pathVariable string
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

func NewWithPathVariable(path, pathVariable string) Endpoint {
	var e = New(path)
	e.pathVariable = strings.Trim(strings.Trim(pathVariable, " "), "/")

	return e
}

func Join(base Endpoint, path string) Endpoint {
	e := New(base.FormattedPath(TrailingSlash) + strings.Trim(strings.Trim(path, " "), "/"))
	return e
}

func JoinWithPathVariable(base Endpoint, path, pathVariable string) Endpoint {
	e := Join(base, pathVariable)
	e.pathVariable = strings.Trim(pathVariable, " ")
	return e
}

func (m Endpoint) Path() string {
	return m.FormattedPath(LeadingSlash)
}

func (m Endpoint) PathWithPathVariable() string {
	var trimmedPathVariable = strings.Trim(strings.Trim(m.pathVariable, " "), "/")

	if trimmedPathVariable == "" {
		return m.Path()
	}

	return m.FormattedPath(LeadingSlash) + "/{" + trimmedPathVariable + "}"
}

func (m Endpoint) FormattedPath(format Format) string {
	var trimmedPath = strings.Trim(m.path, " ")

	if trimmedPath == rootPath {
		return trimmedPath
	}

	return applyFormatting(format, trimmedPath)
}

func (m Endpoint) FormattedPathWithPathVariable(format Format) string {
	var trimmedPathVariable = strings.Trim(strings.Trim(m.pathVariable, " "), "/")

	if trimmedPathVariable == "" {
		return m.FormattedPath(format)
	}

	return applyFormatting(format, m.PathWithPathVariable())

}

func applyFormatting(format Format, str string) string {
	strippedString := strings.Trim(str, "/")
	switch format {
	case LeadingSlash:
		return "/" + strippedString
	case TrailingSlash:
		return strippedString + "/"
	case LeadingAndTrailingSlash:
		return "/" + strippedString + "/"
	case NoSlash:
		return strippedString
	}
	return str
}
