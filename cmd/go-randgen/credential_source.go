package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

type credentialInput struct {
	flagValue string
	flagSet   bool
	filePath  string
	envName   string
	label     string
}

func resolveCredential(input credentialInput) (string, error) {
	envValue, envSet := os.LookupEnv(input.envName)
	sourceCount := 0
	if input.flagSet {
		sourceCount++
	}
	if input.filePath != "" {
		sourceCount++
	}
	if envSet {
		sourceCount++
	}

	if sourceCount == 0 {
		return "", fmt.Errorf("a %s is required", input.label)
	}
	if sourceCount > 1 {
		return "", fmt.Errorf("configure %s using exactly one source", input.label)
	}

	value := input.flagValue
	if input.filePath != "" {
		file, err := os.Open(input.filePath)
		if err != nil {
			return "", fmt.Errorf("read %s file %q: %v", input.label, input.filePath, err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return "", fmt.Errorf("read %s file %q metadata: %v", input.label, input.filePath, err)
		}
		if !fileInfo.Mode().IsRegular() {
			return "", fmt.Errorf("%s file %q is not a regular file", input.label, input.filePath)
		}
		permissions := fileInfo.Mode().Perm()
		if runtime.GOOS != "windows" && permissions != 0400 && permissions != 0600 {
			return "", fmt.Errorf("%s file %q has insecure permissions %04o; use 0400 or 0600", input.label, input.filePath, fileInfo.Mode().Perm())
		}

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("read %s file %q: %v", input.label, input.filePath, err)
		}
		value = strings.TrimRight(string(contents), "\r\n")
	} else if envSet {
		value = envValue
	}

	if value == "" {
		return "", fmt.Errorf("%s source is empty", input.label)
	}
	return value, nil
}

func resolveCredentialList(input credentialInput) ([]string, error) {
	value, err := resolveCredential(input)
	if err != nil {
		return nil, err
	}

	separator := func(r rune) bool { return r == ',' }
	if input.filePath != "" {
		separator = func(r rune) bool { return r == '\n' || r == '\r' }
	}
	values := strings.FieldsFunc(value, separator)
	if len(values) == 0 {
		return nil, fmt.Errorf("%s source is empty", input.label)
	}
	return values, nil
}
