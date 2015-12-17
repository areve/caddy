package setup

import (
	"github.com/mholt/caddy/middleware"
)

func Hide(c *Controller) (middleware.Middleware, error) {
	hide := middleware.Hide{}
	hc := middleware.HideConfig{}

	for c.Next() {
		for true{
			if c.NextBlock() {
				switch c.Val() {
					case "prefix":
						for c.NextArg() {
							hc.Prefix = append(hc.Prefix, c.Val())
						}
					case "suffix":
						for c.NextArg() {
							hc.Suffix = append(hc.Suffix, c.Val())
						}
					case "name":
						for c.NextArg() {
							hc.Name = append(hc.Name, c.Val())
						}
					case "path":
						for c.NextArg() {
							hc.Path = append(hc.Path, c.Val())
						}
					default:
						return nil, c.ArgErr()
				}
			} else if c.NextArg() {
				switch c.Val() {
				case "matchcase":
					hc.MatchCase = true
				default:
					return nil, c.ArgErr()
				}
			} else {
				hide.Configs = append(hide.Configs, hc)
				hc = middleware.HideConfig{}
				break
			}
		}
	}

	c.Hide = hide

	return func(next middleware.Handler) middleware.Handler {
		hide.Next = next
		return hide
	}, nil
}


