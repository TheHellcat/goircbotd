package ircbotint
/*
import (
    "fmt"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
)


/**
 *
 * /
func DmSet( db string, table string, keys []string, kvSet map[string]string ) {
    var sqlCheck string
    var sqlSet string
    var s string
    var t string
    var i int
    var keys map[int]string

    s = ""
    i = 0
    keys = make(map[int]string)
    for t = range keys {
        if len(s) == 0 {
            s = fmt.Sprintf("%s=?", t)
        } else {
            s = fmt.Sprintf(" AND %s=?", t)
        }
        keys[i] = kvSet[t]
        i++
    }
    i--

    db, err := sql.Open("sqlite3", fmt.Sprintf( "%s%s.db", hcIrc.GetDataDir(), db ))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf( "[DATAMANAGERDEBUG] ERROR opening database: %s\n", err.Error() )
            return
        }
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf( "[DATAMANAGERDEBUG] ERROR starting DB session: %s\n", err.Error() )
            return
        }
    }

    sqlCheck = fmt.Sprintf( "SELECT * FROM %s WHERE %s;", table, s )
    stmt, err := tx.Prepare(sqlCheck)
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf( "[DATAMANAGERDEBUG] ERROR setting up statement: %s\n", err.Error() )
            return
        }
    }
    defer stmt.Close()

    rs, err := stmt.Exec(i, fmt.Sprintf("こんにちわ世界%03d", i))
    if err != nil {
        log.Fatal(err)
    }
}
*/