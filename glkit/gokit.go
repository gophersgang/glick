package glkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/documize/glick"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

// MakeEndpoint returns a gokit.io endpoint from a glick library,
// it is intended for use inside servers constructed using gokit.
func MakeEndpoint(l *glick.Library, api, action string) endpoint.Endpoint {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		return l.Run(ctx, api, action, in)
	}
}

// PluginKitJSONoverHTTP enables plugin commands created using gokit.io
func PluginKitJSONoverHTTP(cmdPath string, ppo glick.ProtoPlugOut) glick.Plugger {
	return func(ctx context.Context, in interface{}) (out interface{}, err error) {
		var j, b []byte
		var r *http.Response
		if j, err = json.Marshal(in); err != nil {
			return nil, err
		}
		if r, err = http.Post(cmdPath, "application/json", bytes.NewReader(j)); err != nil {
			return nil, err
		}
		if b, err = ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		}
		out = ppo()
		if err = json.Unmarshal(b, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}

// ConfigKit provides the Configurator for the GoKit class of plugin
func ConfigKit(lib *glick.Library) error {
	return lib.AddConfigurator("KIT", func(l *glick.Library, line int, cfg *glick.Config) error {
		ppo, err := l.ProtoPlugOut(cfg.API)
		if err != nil {
			return fmt.Errorf(
				"entry %d Go-Kit plugin error for api: %s action: %s error: %s",
				line, cfg.API, cfg.Action, err)
		}
		if !cfg.JSON {
			return fmt.Errorf(
				"entry %d Go-Kit: non-JSON plugins are not supported",
				line)
		}
		if err := l.RegPlugin(cfg.API, cfg.Action, PluginKitJSONoverHTTP(cfg.Path, ppo)); err != nil {
			return fmt.Errorf("entry %d Go-Kit register plugin error: %v",
				line, err)
		}
		return nil
	})
}
