package conf

import (
	"context"
	"net/http"
	"sync"

	"github.com/LearnLoop365/flxr-core/db/kvdb"
	"github.com/LearnLoop365/flxr-core/db/sqldb"
)

type Common struct {
	AppName      string          `json:"app_name"`
	AppRoot      string          `json:"-"` // from Compiled Paths
	Listen       string          `json:"listen"`
	Host         string          `json:"host"` // can be used to generate public url endpoints
	Context      context.Context `json:"-"`    // [Interface]
	VolatileKV   *sync.Map       `json:"-"`
	CommonDBConf CommonDBConf    `json:"-"` // Use separate conf file. Make Custom filed to extend
	KVDBClient   kvdb.Client     `json:"-"` // [Interface]
	MainDBClient sqldb.Client    `json:"-"` // [Interface]
	HttpClient   *http.Client    `json:"-"`
	SessionLocks *sync.Map       `json:"-"` // map[string]*sync.Mutex
	DebugOpts    DebugOpts       `json:"debug_opts"`
}

type CommonDBConf struct {
	KV   kvdb.Conf  `json:"kv"`
	Main sqldb.Conf `json:"main"`
}

type DebugOpts struct {
	MaintenanceMode int `json:"maintenance_mode"`
	AuthBreak       int `json:"auth_break"`
}
