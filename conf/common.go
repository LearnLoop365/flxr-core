package conf

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/LearnLoop365/flxr-core/db"
	"github.com/LearnLoop365/flxr-core/db/kvdb"
	"github.com/LearnLoop365/flxr-core/db/sqldb"
)

type Common struct {
	AppName             string               `json:"app_name"`
	AppRoot             string               `json:"-"` // filled from compiled paths
	Listen              string               `json:"listen"`
	Host                string               `json:"host"` // can be used to generate public url endpoints
	Context             context.Context      `json:"-"`
	VolatileKV          *sync.Map            `json:"-"`
	DBConf              CommonDBConf         `json:"-"` // Init manually. e.g. for separate file
	KVDBClient          kvdb.Client          `json:"-"`
	MainDBClient        sqldb.Client         `json:"-"`
	MainDBRawStore      *sqldb.RawStore      `json:"-"`
	MainDBPreparedStore map[string]*sql.Stmt `json:"-"`
	HttpClient          *http.Client         `json:"-"`
	SessionLocks        *sync.Map            `json:"-"`          // map[string]*sync.Mutex
	DebugOpts           DebugOpts            `json:"debug_opts"` // Do not promote
}

func (e *Common) CleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")

	// clean up prepared statements if any
	if e.MainDBPreparedStore != nil {
		for k, stmt := range e.MainDBPreparedStore {
			if stmt != nil {
				if err := stmt.Close(); err != nil {
					log.Printf("[ERROR] closing stmt %s: %v\n", k, err)
				}
			}
		}
		log.Println("[INFO] sql prepared statements closed")
	}

	// clean up DB clients
	db.CloseClient("KVDBClient", e.KVDBClient)
	db.CloseClient("MainDBClient", e.MainDBClient)

	log.Println("[INFO] App Resource Cleanup Complete")
}

type CommonDBConf struct {
	KV   kvdb.Conf  `json:"kv"`
	Main sqldb.Conf `json:"main"`
}

type DebugOpts struct {
	MaintenanceMode int `json:"maintenance_mode"`
	AuthBreak       int `json:"auth_break"`
}
