package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/solidcoredata/scd/api"

	"github.com/google/go-jsonnet"
)

/*
	ResourceNone    ResourceType = ""
	ResourceAuth    ResourceType = "solidcoredata.org/resource/auth"
	ResourceURL     ResourceType = "solidcoredata.org/resource/url"
	ResourceSPACode ResourceType = "solidcoredata.org/resource/spa-code"
	ResourceQuery   ResourceType = "solidcoredata.org/resource/query"

	var LoginState_value = map[string]int32{
		"Missing":        0,
		"Error":          1,
		"None":           2,
		"Granted":        3,
		"U2F":            4,
		"ChangePassword": 5,
	}
*/

type Resource struct {
	Name    string
	Type    string // api.ResourceType
	Parent  string
	Include []string
	C       map[string]interface{}
}

type LoginBundle struct {
	ConsumeRedirect bool
	Resource        string
	Prefix          string
	LoginState      string // api.LoginState_value
}

type Application struct {
	AuthResource string
	Host         []string
	Login        []LoginBundle
}

// ResourceFile represents a donwloadable ll
type ResourceFile struct {
	Name string
	File string
}

type ServiceConfiguration struct {
	Name        string
	Application []Application
	Resource    []Resource
	Files       []ResourceFile
}

func OpenServiceConfiguration(p string) (*api.ServiceBundle, map[string]string, error) {
	vm := jsonnet.MakeVM()

	pabs, err := filepath.Abs(p)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get ABS file path of %q: %v", p, err)
	}
	dir, _ := filepath.Split(pabs)
	vm.Importer(&jsonnet.FileImporter{JPaths: []string{dir}})

	bfile, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, nil, err
	}
	v, err := vm.EvaluateSnippet(p, string(bfile))
	if err != nil {
		return nil, nil, err
	}

	decode := json.NewDecoder(strings.NewReader(v))
	decode.DisallowUnknownFields()
	decode.UseNumber()

	sc := ServiceConfiguration{}
	err = decode.Decode(&sc)
	if err != nil {
		return nil, nil, err
	}
	sb := &api.ServiceBundle{
		Name:        sc.Name,
		Application: make([]*api.ApplicationBundle, 0, len(sc.Application)),
		Resource:    make([]*api.Resource, 0, len(sc.Resource)),
	}
	// Copy over Application and Resource.
	for _, rc := range sc.Resource {
		r := &api.Resource{
			Name:    rc.Name,
			Parent:  rc.Parent,
			Include: rc.Include,
			Type:    rc.Type,
		}

		c := rc.C
		kind := c["Kind"]
		delete(rc.C, "Kind")
		switch kind {
		case "", nil:
			// No configuration.
			if len(c) > 0 {
				return nil, nil, fmt.Errorf("missing kind in configuration")
			}
		default:
			return nil, nil, fmt.Errorf("unknown kind %v", kind)
		case "auth":
			o := &api.ConfigureAuth{}
			o.Environment = c["Environment"].(string)
			area := c["Area"].(string)
			if avalue, found := api.ConfigureAuth_AreaType_value[area]; found {
				o.Area = api.ConfigureAuth_AreaType(avalue)
			} else {
				return nil, nil, fmt.Errorf("unknown area value %q, want one of %+v", area, api.ConfigureAuth_AreaType_value)
			}
			r.Configuration, err = o.Encode()
			if err != nil {
				return nil, nil, err
			}
		case "url":
			o := &api.ConfigureURL{}
			o.MapTo = c["MapTo"].(string)
			switch cc := c["Config"].(type) {
			case map[string]interface{}:
				cbyte, err := json.Marshal(cc)
				if err != nil {
					return nil, nil, err
				}
				o.Config = string(cbyte)
			}
			r.Configuration, err = o.Encode()
			if err != nil {
				return nil, nil, err
			}
		case "spa":
			cbyte, err := json.Marshal(c)
			if err != nil {
				return nil, nil, err
			}
			r.Configuration = cbyte
		}

		sb.Resource = append(sb.Resource, r)
	}
	for _, ac := range sc.Application {
		a := &api.ApplicationBundle{
			AuthConfiguredResource: ac.AuthResource,
			Host:        ac.Host,
			LoginBundle: make([]*api.LoginBundle, 0, len(ac.Login)),
		}
		sb.Application = append(sb.Application, a)

		for _, lc := range ac.Login {
			l := &api.LoginBundle{
				Prefix:          lc.Prefix,
				ConsumeRedirect: lc.ConsumeRedirect,
				Resource:        lc.Resource,
			}
			if ls, found := api.LoginState_value[lc.LoginState]; found {
				l.LoginState = api.LoginState(ls)
			} else {
				return nil, nil, fmt.Errorf("unknown login state %q, need one of %+v", lc.LoginState, api.LoginState_value)
			}
			a.LoginBundle = append(a.LoginBundle, l)
		}
	}

	files := make(map[string]string, len(sc.Files))
	for _, f := range sc.Files {
		b, err := ioutil.ReadFile(filepath.Join(dir, f.File))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read %q: %v", f.Name, err)
		}
		files[f.Name] = string(b)
	}

	return sb, files, nil
}
