package main

import (
	"fmt"
	"strconv"
	"strings"
)

type serverAddress struct {
	host string
	port int
}

func (sa *serverAddress) String() string {
	return fmt.Sprintf("%s:%d", sa.host, sa.port)
}

func (sa *serverAddress) Set(flagValue string) error {
	fValue := strings.Split(flagValue, ":")
	if len(fValue) != 2 {
		return fmt.Errorf("invalid server address format: %s", flagValue)
	}

	var err error
	sa.port, err = strconv.Atoi(fValue[1])
	if err != nil {
		return fmt.Errorf("invalid server address port: %s", fValue[1])
	}

	if fValue[0] != "" {
		sa.host = fValue[0]
	}

	return nil
}
