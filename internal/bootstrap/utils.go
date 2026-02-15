package bootstrap

import (
	"fmt"
	"gateway/config"
	"gateway/internal/common"
	"strings"
)

func checkProxyRoutes(cfg config.RouterSettings, reserved ...string) error {
	set := common.NewSet(reserved...)

	for _, route := range cfg.Routes {
		for _, path := range route.Paths {
			if set.Has(strings.TrimSuffix(path.Path, "/")) {
				return fmt.Errorf(
					"proxy rule contain reserved endpoint: %s%s",
					route.Host, path.Path,
				)
			}
		}
	}

	return nil
}
