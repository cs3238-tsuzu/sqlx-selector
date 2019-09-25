package sqlxselect

import (
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
	"golang.org/x/xerrors"
)

type SqlxSelector struct {
	node    *structElementNode
	columns []string
	Errors  []error
}

func New(dst interface{}) (*SqlxSelector, error) {
	return NewWithMapper(dst, reflectx.NewMapperFunc("db", strings.ToLower))
}

func NewWithMapper(dst interface{}, mapper *reflectx.Mapper) (*SqlxSelector, error) {
	m := mapper.FieldMap(reflect.ValueOf(dst))

	node := &structElementNode{}

	for path := range m {
		node.addChild(splitPath(path)...)
	}

	return &SqlxSelector{
		node: node,
	}, nil
}

func (s *SqlxSelector) Select(column string) *SqlxSelector {
	s.columns = append(s.columns, "`"+column+"`")

	return s
}

func (s *SqlxSelector) SelectAs(column, as string) *SqlxSelector {
	s.columns = append(s.columns, "`"+column+"` AS "+doubleQuote(as))

	return s
}

func (s *SqlxSelector) SelectStruct(column string, limit ...string) *SqlxSelector {
	return s.SelectStructAs(column, column, limit...)
}

func (s *SqlxSelector) SelectStructAs(column, as string, limit ...string) *SqlxSelector {
	ass := splitPath(as)

	if len(ass) != 0 && ass[len(ass)-1] == "*" {
		ass = ass[:len(ass)-1]
	}

	node := s.node.findNode(ass...)

	if node == nil {
		s.Errors = append(s.Errors, xerrors.Errorf("unknown node in %v", as))
		return s
	}

	columnPrefix := strings.TrimSuffix(column, "*")

	elms := node.listElements()

	check := true
	if len(limit) == 0 {
		check = false
		limit = elms
	}

	elmsSet := map[string]struct{}{}
	if check {
		for i := range elms {
			elmsSet[elms[i]] = struct{}{}
		}
	}

	for i := range limit {
		if check {
			_, found := elmsSet[limit[i]]

			if !found {

			}
		}

		s.SelectAs(columnPrefix+limit[i], strings.Join(append(ass, limit[i]), "."))
	}

	return s
}

func (s *SqlxSelector) String() string {
	if len(s.Errors) != 0 {
		return ""
	}
	return strings.Join(s.columns, ",")
}

func (s *SqlxSelector) StringWithError() (string, error) {
	if len(s.Errors) != 0 {
		return "", flattenErrors(s.Errors)
	}
	return strings.Join(s.columns, ","), nil
}
