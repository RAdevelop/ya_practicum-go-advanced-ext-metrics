package agent

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerAddress struct {
	Host string
	Port int
}

func (sa *ServerAddress) String() string {
	return fmt.Sprintf("http://%s:%d", sa.Host, sa.Port)
}

func (sa *ServerAddress) Set(flagValue string) error {

	if strings.Contains(flagValue, "//") {
		return fmt.Errorf("invalid server address format: %s, set host:port without schema", flagValue)
	}

	fValue := strings.Split(flagValue, ":")
	if len(fValue) != 2 {
		return fmt.Errorf("invalid server address format: %s, set host:port without schema", flagValue)
	}

	var err error
	sa.Port, err = strconv.Atoi(fValue[1])
	if err != nil {
		return fmt.Errorf("invalid server address port: %s", fValue[1])
	}

	if fValue[0] != "" {
		sa.Host = fValue[0]
	}

	return nil
}
