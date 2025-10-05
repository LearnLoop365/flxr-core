package storages

import "github.com/LearnLoop365/flxr-core/storages/keystores"

type Conf struct {
	KeyStoreConf keystores.Conf `json:"key_store"`
}
