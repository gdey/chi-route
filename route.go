package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

const (
	tagName = "query"
)

// Token represents a go-chi token in a url
type Token string

func (tok Token) AsInt32(r *http.Request) (int32, error) {
	i64, err := strconv.ParseInt(chi.URLParam(r, tok.String()), 10, 32)
	return int32(i64), err
}
func (tok Token) AsInt64(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, tok.String()), 10, 64)
}
func (tok Token) Get(r *http.Request) string { return chi.URLParam(r, tok.String()) }

func (tok Token) String() string {
	// we need to see if there is a ":".
	idx := strings.IndexByte(string(tok), ':')
	if idx == -1 {
		return string(tok)
	}
	// we need to return everthing before index
	return string(tok[:idx])
}

func (tok Token) Token() string {
	return "{" + string(tok) + "}"
}

// Pattern builds route strings that match the chi URL pattern
func Pattern(elements ...interface{}) string {
	return Join("/", elements...)
}

func Join(sep string, elements ...interface{}) string {
	strElements := make([]string, 1, len(elements)+1)
	for _, e := range elements {
		switch el := e.(type) {
		case Token:
			strElements = append(strElements, el.Token())
		case string:
			strElements = append(strElements, el)
		case fmt.Stringer:
			strElements = append(strElements, el.String())
		default:
			strElements = append(strElements, fmt.Sprintf("%v", el))
		}
	}
	return strings.Join(strElements, sep)
}
