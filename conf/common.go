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
	AppRoot      string          `json:"-"` // filled from compiled paths
	Listen       string          `json:"listen"`
	Host         string          `json:"host"` // can be used to generate public url endpoints
	Context      context.Context `json:"-"`
	VolatileKV   *sync.Map       `json:"-"`
	DBConf       CommonDBConf    `json:"-"` // Init manually. e.g. for separate file
	KVDBClient   kvdb.Client     `json:"-"`
	MainDBClient sqldb.Client    `json:"-"`
	HttpClient   *http.Client    `json:"-"`
	SessionLocks *sync.Map       `json:"-"`          // map[string]*sync.Mutex
	DebugOpts    DebugOpts       `json:"debug_opts"` // Do not promote
}

type CommonDBConf struct {
	KV   kvdb.Conf  `json:"kv"`
	Main sqldb.Conf `json:"main"`
}

type DebugOpts struct {
	MaintenanceMode int `json:"maintenance_mode"`
	AuthBreak       int `json:"auth_break"`
}
