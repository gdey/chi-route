package route_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gdey/go-route"
)

type getter interface {
	Get(string) string
}
type getterMap map[string]string

func (gm getterMap) Get(key string) string { return gm[key] }

func (gm getterMap) Set(key, value string) { gm[key] = value }

type qparams struct {
	Width    int      `query:"w"`
	Height   int      `query:"h"`
	MaxZoom  *float64 `query:"max_zoom"`
	Provider string   `query:"provider"`
	MapID    string   `query:"map_id"`
	MapUser  string   `query:"map_user"`
}

func (qp qparams) String() string {
	maxz := 0.0
	if qp.MaxZoom != nil {
		maxz = *qp.MaxZoom
	}
	return fmt.Sprintf("dim: %dx%d : %v, %s %s %s", qp.Width, qp.Height, maxz, qp.Provider, qp.MapID, qp.MapUser)
}

func (qparams) cmp(x, y interface{}) bool {
	q1, q2 := x.(qparams), y.(*qparams)
	if !(q1.Width == q2.Width && q1.Height == q2.Height) {
		return false
	}
	success := true
	if q1.MaxZoom == nil || q2.MaxZoom == nil {
		success = success && q1.MaxZoom == q2.MaxZoom
	} else {
		success = success && *q1.MaxZoom == *q2.MaxZoom
	}
	success = success && q1.MapUser == q2.MapUser
	success = success && q1.MapID == q2.MapID
	success = success && q1.Provider == q2.Provider
	return success
}

func TestParseQuery(t *testing.T) {

	type tcase struct {
		Struct   interface{}
		Expected interface{}
		Params   getter
		Cmp      func(x, y interface{}) bool
		UseErrIs bool
		Err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		var errCheck = errors.As
		if tc.UseErrIs {
			errCheck = func(err error, target interface{}) bool { return errors.Is(err, target.(error)) }
		}
		if tc.Cmp == nil {
			tc.Cmp = reflect.DeepEqual
		}
		return func(t *testing.T) {
			gotErr := route.ParseQuery(tc.Params, tc.Struct)
			if tc.Err != nil {
				if !errCheck(gotErr, tc.Err) {
					t.Errorf("error, expected %v, got %v", tc.Err, gotErr)
				}
				return
			}
			if gotErr != nil && tc.Err != nil {
				t.Errorf("error, expected nil, got %v", gotErr)
				return
			}
			if !tc.Cmp(tc.Expected, tc.Struct) {
				t.Errorf("struct, expected %v got %v", tc.Expected, tc.Struct)
				return
			}
		}
	}

	maxZoom := 2.4

	tests := map[string]tcase{
		"base": tcase{
			Params: getterMap{
				"w":        "10",
				"h":        "10",
				"max_zoom": "2.4",
			},
			Expected: qparams{
				Width:   10,
				Height:  10,
				MaxZoom: &maxZoom,
			},
			Struct: &qparams{},
			Cmp:    qparams{}.cmp,
		},
		"base w defaults": tcase{
			Params: getterMap{
				"w":        "10",
				"h":        "10",
				"max_zoom": "2.4",
			},
			Expected: qparams{
				Width:    10,
				Height:   10,
				MaxZoom:  &maxZoom,
				Provider: "google",
			},
			Struct: &qparams{
				Width:    100,
				Height:   100,
				Provider: "google",
			},
			Cmp: qparams{}.cmp,
		},
		"embbed structs": func() tcase {
			type Size struct {
				Width  int `query:"w"`
				Height int `query:"h"`
			}
			type MapProvider struct {
				Provider string `query:"provider"`
				Id       string `query:"map_id"`
				User     string `query:"map_user"`
			}
			type QParams struct {
				Size
				MapProvider
			}

			return tcase{
				Params: getterMap{
					"w":        "10",
					"h":        "10",
					"provider": "google",
				},
				Expected: &QParams{
					Size: Size{
						Width:  10,
						Height: 10,
					},
					MapProvider: MapProvider{
						Provider: "google",
					},
				},
				Struct: &QParams{},
				Cmp:    reflect.DeepEqual,
			}

		}(),
		"mixed embbed struct": func() tcase {
			type MapProvider struct {
				Provider string `query:"provider"`
				Id       string `query:"map_id"`
				User     string `query:"map_user"`
			}
			type QParams struct {
				Width  int `query:"w"`
				Height int `query:"h"`
				MapProvider
			}

			return tcase{
				Params: getterMap{
					"w":        "10",
					"h":        "10",
					"provider": "google",
				},
				Expected: &QParams{
					Width:  10,
					Height: 10,
					MapProvider: MapProvider{
						Provider: "google",
					},
				},
				Struct: &QParams{},
				Cmp:    reflect.DeepEqual,
			}

		}(),
		"multi embbed structs": func() tcase {
			type MapUser struct {
				Id   string `query:"map_id"`
				User string `query:"map_user"`
			}
			type MapProvider struct {
				Provider string `query:"provider"`
				User     MapUser
			}
			type QParams struct {
				Width  int `query:"w"`
				Height int `query:"h"`
				MapProvider
			}

			return tcase{
				Params: getterMap{
					"w":        "10",
					"h":        "10",
					"provider": "google",
					"map_user": "gdey",
				},
				Expected: &QParams{
					Width:  10,
					Height: 10,
					MapProvider: MapProvider{
						Provider: "google",
						User: MapUser{
							User: "gdey",
						},
					},
				},
				Struct: &QParams{},
				Cmp:    reflect.DeepEqual,
			}

		}(),
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestCreateQuery(t *testing.T) {
	type tcase struct {
		Struct   interface{}
		Expected getter
		Err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			// We are going to set values into getter
			var params = make(getterMap)
			err := route.CreateQuery(tc.Struct, params)
			if tc.Err != nil {
				if err == nil {
					t.Errorf("error, expected %v, got nil", tc.Err)
					return
				}
				if !errors.As(err, &tc.Err) {
					t.Errorf("error, expected %v, got %v", tc.Err, err)
					return
				}
				return
			}

			if !reflect.DeepEqual(params, tc.Expected) {
				t.Errorf("values, expected %v, got %v", tc.Expected, params)
				return
			}

		}
	}

	maxZoom := 2.4
	tests := map[string]tcase{
		"base": tcase{
			Expected: getterMap{
				"w":        "10",
				"h":        "10",
				"max_zoom": "2.4",
			},
			Struct: &qparams{
				Width:   10,
				Height:  10,
				MaxZoom: &maxZoom,
			},
		},
		"max zoom nil": tcase{
			Expected: getterMap{
				"w": "10",
				"h": "10",
			},
			Struct: &qparams{
				Width:  10,
				Height: 10,
			},
		},
		"multi pointer": tcase{
			Expected: getterMap{
				"h": "5",
			},
			Struct: struct {
				val **int `query:"h"`
			}{
				val: func() **int {
					v := int(5)
					val := &v
					return &val
				}(),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}
