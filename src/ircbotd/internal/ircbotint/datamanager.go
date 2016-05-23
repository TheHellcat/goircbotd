package ircbotint

import (
    "fmt"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)


var dmCacheCheckTable map[string]string


/**
 *
 */
func DmCheckTable( database string, table string, createStmt string ) {
    var query string
    var err error
    var needCreate bool
    var rName string

    query = fmt.Sprintf( "SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", table )

    db, err := sql.Open("sqlite3", fmt.Sprintf("%s%s.db", hcIrc.GetDataDir(), database))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] ERROR opening database: %s\n", err.Error())
        }
        return
    }
    defer db.Close()

    rs, err := db.Query( query )

    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] Error looking up table '%s.%s: %s'\n", database, table, err.Error())
        }
        return
    }
    if rs == nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] Error looking up table '%s.%s: Got a NIL result set'\n", database, table)
        }
        return
    }

    needCreate = false
    if !rs.Next() {
        needCreate = true
    } else {
        rs.Scan( &rName )
        if rName != table {
            needCreate = true
        }
    }

    if needCreate {
        // table does not exist, execute the supplied create query to fix this
        _, err = db.Exec( createStmt )
        if err != nil {
            if hcIrc.Debugmode {
                fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] ERROR creating table '%s.%s': %s\n", database, table, err.Error())
            }
        } else {
            if hcIrc.Debugmode {
                fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] Created table '%s.%s'\n", database, table)
            }
        }
    }
}


/**
 *
 */
func DmSet( database string, table string, keys []string, kvSet map[string]string ) {
    var sqlCheck string
    //var sqlSet string
    var s string
    var t string
    var i int
    var keyVs map[int]string
    //var isUpdate bool

    s = ""
    i = 0
    keyVs = make(map[int]string)
    for _, t = range keys {
        if len(s) == 0 {
            s = fmt.Sprintf("%s=?", t)
        } else {
            s = fmt.Sprintf(" AND %s=?", t)
        }
        keyVs[i] = kvSet[t]  // build map with values in the same order as keys
        i++
    }
    i--

    // opem DB connection

    db, err := sql.Open("sqlite3", fmt.Sprintf("%s%s.db", hcIrc.GetDataDir(), database))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR opening database: %s\n", err.Error())
            return
        }
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR starting DB session: %s\n", err.Error())
            return
        }
    }


    // check if we need to INSERT or UPDATE

    sqlCheck = fmt.Sprintf("SELECT * FROM %s WHERE %s;", table, s)
    stmt, err := tx.Prepare(sqlCheck)
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR setting up statement: %s\n", err.Error())
            return
        }
    }
    defer stmt.Close()

    // super fugly workaround, till I figured out how to make this dynamic
    if len(keyVs) == 1 {
        rs, err := stmt.Query(keyVs[0])
        err = err
        fmt.Println(rs)
    } else if len(keys) == 2 {
        rs, err := stmt.Query(keyVs[0], keyVs[1])
        err = err
        fmt.Println(rs)
    } else if len(keys) == 3 {
        rs, err := stmt.Query(keyVs[0], keyVs[1], keyVs[2])
        err = err
        fmt.Println(rs)
    } else if len(keys) == 4 {
        rs, err := stmt.Query(keyVs[0], keyVs[1], keyVs[2], keyVs[3])
        err = err
        fmt.Println(rs)
    } else {
        //err = error( fmt.Sprintf( "Unsupportet number of keys given: %d", len(keyVs)) )
    }
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR executing statement: %s\n", err.Error())
            return
        }
    }
    tx.Commit()

}
