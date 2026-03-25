package plugin

// This file embeds the sqlite_protobuf shared library directly into the plugin
// binary and loads it into every SQLite connection via mattn/go-sqlite3.
//
// The .so is written to a temp file once at startup (dlopen requires a real
// filesystem path) and reused for every subsequent connection.

import (
	"database/sql"
	_ "embed"
	"os"
	"sync"

	sqlite3 "github.com/mattn/go-sqlite3"
)

//go:embed libsqlite_protobuf.so
var protobufExtBytes []byte

var (
	extOnce sync.Once
	extPath string
)

func loadedExtPath() string {
	extOnce.Do(func() {
		f, err := os.CreateTemp("", "libsqlite_protobuf_*.so")
		if err != nil {
			panic("sqlite_protobuf: create temp file: " + err.Error())
		}
		if _, err := f.Write(protobufExtBytes); err != nil {
			panic("sqlite_protobuf: write temp file: " + err.Error())
		}
		if err := os.Chmod(f.Name(), 0755); err != nil {
			panic("sqlite_protobuf: chmod temp file: " + err.Error())
		}

		f.Close()
		extPath = f.Name()
	})
	return extPath
}

func init() {
	sql.Register("sqlite", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.LoadExtension(loadedExtPath(), "sqlite3_extension_init")
		},
	})
}
