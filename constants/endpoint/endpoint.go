package endpoint

import "strings"

const rootPath = "/"

var Root = New(rootPath)

type Format int64

const (
	LeadingSlash Format = iota
	TrailingSlash
	LeadingAndTrailingSlash
	NoSlash
)

type Endpoint struct {
	path          string
	pathVariables []string
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

func NewWithPathVariable(path string, pathVariable ...string) Endpoint {
	var e = New(path)
	e.pathVariables = sanitizeArgs(pathVariable)

	return e
}

func Join(base Endpoint, path string) Endpoint {
	e := New(base.FormattedPath(TrailingSlash) + strings.Trim(strings.ReplaceAll(path, " ", ""), "/"))
	e.pathVariables = base.pathVariables
	return e
}

func JoinWithPathVariable(base Endpoint, path string, pathVariables ...string) Endpoint {
	e := Join(base, path)
	basePathVariables := base.pathVariables
	newPathVariables := sanitizeArgs(pathVariables)
	e.pathVariables = append(basePathVariables, newPathVariables...)
	return e
}

func (m Endpoint) Path() string {
	return m.FormattedPath(LeadingSlash)
}

func (m Endpoint) PathWithPathVariable() string {
	var combinedPathVariables = strings.Join(m.pathVariables, "/")

	if combinedPathVariables == "" {
		return m.Path()
	}

	return m.FormattedPath(LeadingSlash) + "/" + combinedPathVariables
}

func (m Endpoint) FormattedPath(format Format) string {
	var trimmedPath = strings.ReplaceAll(m.path, " ", "")

	if trimmedPath == rootPath {
		return trimmedPath
	}

	return applyFormatting(format, trimmedPath)
}

func (m Endpoint) FormattedPathWithPathVariable(format Format) string {
	var trimmedPathVariable = strings.Join(m.pathVariables, "/")

	if trimmedPathVariable == "" {
		return m.FormattedPath(format)
	}

	return applyFormatting(format, m.PathWithPathVariable())

}

func sanitizeArgs(pathVariable []string) []string {
	pathVariables := make([]string, len(pathVariable))
	for index, val := range pathVariable {
		trimmedVal :=  strings.Trim(strings.ReplaceAll(val, " ", ""), "/")
		if trimmedVal != "" {
			pathVariables[index] = "{" + trimmedVal + "}"
		}
	}

	return pathVariables
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
