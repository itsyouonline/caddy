package caddyhttp

import (
	// plug in the server
	_ "github.com/itsyouonline/caddy/caddyhttp/httpserver"

	// plug in the standard directives
	_ "github.com/itsyouonline/caddy/caddyhttp/basicauth"
	_ "github.com/itsyouonline/caddy/caddyhttp/bind"
	_ "github.com/itsyouonline/caddy/caddyhttp/browse"
	_ "github.com/itsyouonline/caddy/caddyhttp/errors"
	_ "github.com/itsyouonline/caddy/caddyhttp/expvar"
	_ "github.com/itsyouonline/caddy/caddyhttp/extensions"
	_ "github.com/itsyouonline/caddy/caddyhttp/fastcgi"
	_ "github.com/itsyouonline/caddy/caddyhttp/gzip"
	_ "github.com/itsyouonline/caddy/caddyhttp/header"
	_ "github.com/itsyouonline/caddy/caddyhttp/index"
	_ "github.com/itsyouonline/caddy/caddyhttp/internalsrv"
	_ "github.com/itsyouonline/caddy/caddyhttp/limits"
	_ "github.com/itsyouonline/caddy/caddyhttp/log"
	_ "github.com/itsyouonline/caddy/caddyhttp/markdown"
	_ "github.com/itsyouonline/caddy/caddyhttp/mime"
	_ "github.com/itsyouonline/caddy/caddyhttp/pprof"
	_ "github.com/itsyouonline/caddy/caddyhttp/proxy"
	_ "github.com/itsyouonline/caddy/caddyhttp/push"
	_ "github.com/itsyouonline/caddy/caddyhttp/redirect"
	_ "github.com/itsyouonline/caddy/caddyhttp/requestid"
	_ "github.com/itsyouonline/caddy/caddyhttp/rewrite"
	_ "github.com/itsyouonline/caddy/caddyhttp/root"
	_ "github.com/itsyouonline/caddy/caddyhttp/status"
	_ "github.com/itsyouonline/caddy/caddyhttp/templates"
	_ "github.com/itsyouonline/caddy/caddyhttp/timeouts"
	_ "github.com/itsyouonline/caddy/caddyhttp/websocket"
	_ "github.com/itsyouonline/caddy/startupshutdown"
)
