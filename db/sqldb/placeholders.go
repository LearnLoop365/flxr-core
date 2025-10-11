package sqldb

var PlaceholderPrefixForDBType = map[string]byte{
	"mysql":  '?',
	"pgsql":  '$',
	"mssql":  '@',
	"oracle": ':',
	"sqlite": 0, // NOTE: sqlite supports all of them
}
