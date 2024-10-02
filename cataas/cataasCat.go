package cataas

import (
    /*
    "io"
    "os"
    */

    "net/http"
    "log"
)

// Filters you can apply to the cat query and its results.
type CatFilters struct {
    Size	string // square, medium, small, xsmall
    Filter	string // mono, negate, custom
    Fit		string // cover, contain, fill, inside, outside
    Position	string // top, right top, right, right bottom, bottom, left bottom, left, left top, center
    Width	int
    Height	int
    Blur	int
    RedHue	int
    GreenHue	int
    BlueHue	int
    Brightness	float32
    Saturation	float32
    Hue		int
    HTML	bool
    JSON	bool
    Lightness	int
}

type BCat interface { // Here for forward compatibility only
    Fetch()	*http.Response
}
type CataasEndpoint interface {
    Fetch()	*http.Response
}

func (c Cat) Fetch() *http.Response {
    if c.Tag {

    }
    apiResponse, err := http.Get("https://cataas.com/cat")
    if err != nil {
	log.Print("-- Failed to fetch cat; error: ", err.Error())
    }
    defer apiResponse.Body.Close()

    return apiResponse
}

type Cat struct {
    Filters CatFilters
    Tag	    string
    Id	    string
    Says    string
}

func NewCat() *Cat {
    c := Cat {}
    return &c
}

func (c Cat) WithFilters(filters CatFilters) *Cat {
    c.Filters = filters
    return &c
}

func (c Cat) WithTag(tag string) *Cat {
    c.Tag = tag
    return &c
}

func (c Cat) WithId(id string) *Cat { // I don't know why the API expects a string for the ID
    c.Id = id
    return &c
}

func (c Cat) Say(caption string) *Cat {
    c.Says = caption
    return &c
}
