package snclient

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"pkg/utils"
)

type CheckWrap struct {
	noCopy        noCopy
	snc           *Agent
	name          string
	commandString string
	config        *ConfigSection
	wrapped       bool
}

func (l *CheckWrap) Build() *CheckData {
	return &CheckData{
		name:            l.name,
		hasInventory:    ScriptsInventory,
		argsPassthrough: true,
	}
}

// CheckWrap wraps existing scripts created by the ExternalScriptsHandler
func (l *CheckWrap) Check(ctx context.Context, snc *Agent, check *CheckData, _ []Argument) (*CheckResult, error) {
	l.snc = snc

	macros := map[string]string{}
	for i := range check.rawArgs {
		macros[fmt.Sprintf("ARG%d", i+1)] = check.rawArgs[i]
	}

	var command string
	if l.wrapped {
		// substitute $ARGn$ in the command
		commandString := ReplaceRuntimeMacros(l.commandString, macros)
		cmdToken := utils.Tokenize(commandString)
		macros["SCRIPT"] = cmdToken[0]
		macros["ARGS"] = strings.Join(cmdToken[1:], " ")
		ext := strings.TrimPrefix(filepath.Ext(cmdToken[0]), ".")
		log.Debugf("command wrapping for extension: %s", ext)
		wrapping, ok := snc.Config.Section("/settings/external scripts/wrappings").GetString(ext)
		if !ok {
			return nil, fmt.Errorf("no wrapping found for extension: %s", ext)
		}
		log.Debugf("%s wrapper: %s", ext, wrapping)
		command = ReplaceRuntimeMacros(wrapping, macros)
	} else {
		macros["ARGS"] = strings.Join(check.rawArgs, " ")
		macros["ARGS\""] = strings.Join(func(arr []string) []string {
			quoteds := make([]string, len(arr))
			for i, v := range arr {
				quoteds[i] = fmt.Sprintf("%q", v)
			}

			return quoteds
		}(check.rawArgs), " ")
		log.Debugf("command before macros expanded: %s", l.commandString)
		command = ReplaceRuntimeMacros(l.commandString, macros)
	}

	// set default timeout
	timeoutSeconds, ok, err := l.config.GetInt("timeout")
	if err != nil || !ok {
		timeoutSeconds = 60
	}

	stdout, stderr, exitCode, _ := l.snc.runExternalCheckString(ctx, command, timeoutSeconds)
	if stderr != "" {
		if stdout != "" {
			stdout += "\n"
		}
		stdout += "[" + stderr + "]"
	}

	return &CheckResult{
		State:  exitCode,
		Output: stdout,
	}, nil
}
