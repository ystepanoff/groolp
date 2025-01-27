package cli

import (
	"fmt"

	"github.com/ystepanoff/groolp/core"
)

func InitConfig() (*core.Config, error) {
	config, err := core.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading tasks config: %v", err)
	}

	return config, nil
}
