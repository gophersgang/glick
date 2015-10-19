package glick

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sync"

	"golang.org/x/net/context"
)

var cmdmtx sync.Mutex // ensure we only run one command at a time, across the system

// PluginCmd only works with an api with a simple Text/Text signature.
// it runs the given operating system command using the input string
// as stdin and putting stdout into the output string.
// At present, to limit stress on system resources,
// only one os command can run at a time via this plugin sub-system.
func PluginCmd(cmdPath string, args []string, model interface{}) Plugin {
	cmdPath, e := exec.LookPath(cmdPath)
	if e != nil {
		return nil
	}
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		var err error
		cmdmtx.Lock()
		defer cmdmtx.Unlock()
		ecmd := exec.Command(cmdPath, args...)
		ecmd.Stdin, err = TextReader(in)
		if err != nil {
			return nil, err
		}
		var outBuf, errBuf bytes.Buffer
		ecmd.Stdout, ecmd.Stderr = &outBuf, &errBuf
		err = ecmd.Start()
		if err != nil {
			return nil, err
		}
		over := make(chan error, 1)
		go func() {
			over <- ecmd.Wait()
		}()
		select {
		case err = <-over:
			if err != nil {
				return nil, err
			}
			return TextConvert(outBuf.Bytes(), model)
		case <-ctx.Done():
			ke := ""
			if runtime.GOOS != "windows" { // Process is not available on windows
				err = ecmd.Process.Kill()
				if err != nil {
					ke = ", kill error: " + err.Error()
				}
			}
			return nil, errors.New("Cmd cancelled via context" + ke)
		}
	}
}

// ConfigCmd provides the Configurator for plugins that run operating system commands.
func ConfigCmd(lib *Library) error {
	return lib.AddConfigurator("CMD", func(l *Library, line int, cfg *Config) error {
		if !(IsText(l.apim[cfg.API].ppi) && IsText(l.apim[cfg.API].ppo())) {
			return fmt.Errorf("entry %d API %s is not of simple type (string/*string) ",
				line, cfg.API)
		}
		pi := PluginCmd(cfg.Path, cfg.Args, l.apim[cfg.API].ppo())
		if err := l.RegPlugin(cfg.API, cfg.Action, pi); err != nil {
			return fmt.Errorf("entry %d CMD register plugin error: %v",
				line, err)
		}
		return nil
	})
}