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
    var db *sql.DB
    var rs *sql.Rows
    var b bool

    _, b = dmCacheCheckTable[table]
    if b {
        // we already checked for this table, so it exists, nothing to do now
        return
    }

    query = fmt.Sprintf( "SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", table )

    db, err = sql.Open("sqlite3", fmt.Sprintf("%s%s.db", hcIrc.GetDataDir(), database))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] ERROR opening database: %s\n", err.Error())
        }
        return
    }
    defer db.Close()

    rs, err = db.Query( query )
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmCheckTable] Error looking up table '%s.%s: %s'\n", database, table, err.Error())
        }
        return
    }
    defer rs.Close()
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
    rs.Close()

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

    db.Close()
    dmCacheCheckTable[table] = table
}


/**
 *
 */
func DmSet( database string, table string, keys []string, kvSet map[string]string ) {
    var vCheck map[string]string
    var sqlSet string
    var s string
    var t, u, v string
    var i, j int
    var keyVs map[int]string
    var isUpdate bool
    var err error
    var kWhere string
    var vWhere map[int]string
    var db *sql.DB
    var tx *sql.Tx
    var stmt *sql.Stmt

    s = ""
    i = 0
    keyVs = make(map[int]string)
    vWhere = make(map[int]string)
    for _, t = range keys {
        if len(s) == 0 {
            s = fmt.Sprintf("%s=?", t)
        } else {
            s = fmt.Sprintf("%s AND %s=?", s, t)
        }
        keyVs[i] = kvSet[t]  // build map with values in the same order as keys
        i++
    }
    i--
    kWhere = s
    vWhere = keyVs

    // check if the entry already exists and we need to UPDATE or INSERT
    vCheck = make(map[string]string)
    for _, s = range keys {
        vCheck[s] = kvSet[s]
    }
    _, i = DmGet( database, table, keys, vCheck )
    if i < 0 {
        isUpdate = false
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] Value does not exist, using INSERT\n")
        }
    } else {
        isUpdate = true
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] Value exists, using UPDATE\n")
        }
    }

    // open DB connection
    db, err = sql.Open("sqlite3", fmt.Sprintf("%s%s.db", hcIrc.GetDataDir(), database))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR opening database: %s\n", err.Error())
        }
        return
    }
    defer db.Close()

    // begin new transaction
    tx, err = db.Begin()
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR starting read transaction: %s\n", err.Error())
        }
        return
    }
    defer tx.Commit()


    // build the actual SQL to write the values into the DB

    keyVs = make(map[int]string)
    if isUpdate {
        s = ""
        i = 0
        for u, t = range kvSet {
            if len(s) == 0 {
                s = fmt.Sprintf("%s=?", u)
            } else {
                s = fmt.Sprintf("%s, %s=?", s, u)
            }
            keyVs[i] = t  // build map with values in the same order as columns
            i++
        }
        i--
        sqlSet = fmt.Sprintf( "UPDATE %s SET %s WHERE %s;", table, s, kWhere )
    } else {
        s = ""
        i = 0
        for u, t = range kvSet {
            if len(s) == 0 {
                s = fmt.Sprintf("%s", u)
                v = "?"
            } else {
                s = fmt.Sprintf("%s, %s", s, u)
                v = fmt.Sprintf("%s, ?", v)
            }
            keyVs[i] = t  // build map with values in the same order as columns
            i++
        }
        i--
        sqlSet = fmt.Sprintf( "INSERT INTO %s (%s) VALUES (%s);", table, s, v )
    }

    // new transaction
    tx, err = db.Begin()
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR starting write transaction: %s\n", err.Error())
        }
        return
    }
    defer tx.Commit()

    stmt, err = tx.Prepare(sqlSet)
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR setting up write statement: %s\n", err.Error())
        }
        return
    }
    defer stmt.Close()

    if hcIrc.Debugmode {
        fmt.Printf("[DATAMANAGERDEBUG][DmSet] Set of %d values to be written\n", len(keyVs))
    }

    // add the WHERE values to our value array/map if we're doing an UPDATE
    if isUpdate {
        j = len(keyVs)
        for i = 0; i < len(vWhere); i++ {
            keyVs[j] = vWhere[i]
            j++
        }
    }

    // same fuglieness as above
    if len(keyVs) == 1 {
        _, err = stmt.Exec(keyVs[0])
    } else if len(keyVs) == 2 {
        _, err = stmt.Exec(keyVs[0], keyVs[1])
    } else if len(keyVs) == 3 {
        _, err = stmt.Exec(keyVs[0], keyVs[1], keyVs[2])
    } else if len(keyVs) == 4 {
        _, err = stmt.Exec(keyVs[0], keyVs[1], keyVs[2], keyVs[3])
    } else if len(keyVs) == 5 {
        _, err = stmt.Exec(keyVs[0], keyVs[1], keyVs[2], keyVs[3], keyVs[4])
    } else if len(keyVs) == 6 {
        _, err = stmt.Exec(keyVs[0], keyVs[1], keyVs[2], keyVs[3], keyVs[4], keyVs[5])
    } else if len(keyVs) == 7 {
        _, err = stmt.Exec(keyVs[0], keyVs[1], keyVs[2], keyVs[3], keyVs[4], keyVs[5], keyVs[6])
    } else {
        //err = error( fmt.Sprintf( "Unsupported number of keys given: %d", len(keyVs)) )
    }

    // close everything off and good bye
    tx.Commit()
    stmt.Close()
    db.Close()

    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmSet] ERROR executing write statement: %s\n", err.Error())
        }
        return
    }
}


/**
 *
 */
func DmGet( database string, table string, getColumns []string, getCriteria map[string]string ) (map[int]map[string]string, int) {
    var sqlCheck string
    var s string
    var t, u string
    var i, j int
    var keyVs map[int]string
    var rs *sql.Rows
    var err error
    var kWhere string
    var db *sql.DB
    var tx *sql.Tx
    var stmt *sql.Stmt
    var cols []string
    var numCols int
    var values []string
    var returnValues map[string]string
    var returnData map[int]map[string]string

    returnData = make(map[int]map[string]string)

    s = ""
    i = 0
    keyVs = make(map[int]string)
    for u, t = range getCriteria {
        if len(s) == 0 {
            s = fmt.Sprintf("%s=?", u)
        } else {
            s = fmt.Sprintf("%s AND %s=?", s, u)
        }
        keyVs[i] = t  // build map with values in the same order as keys
        i++
    }
    i--

    // open DB connection
    db, err = sql.Open("sqlite3", fmt.Sprintf("%s%s.db", hcIrc.GetDataDir(), database))
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmGet] ERROR opening database: %s\n", err.Error())
        }
        return nil, -1
    }
    defer db.Close()

    // begin new transaction
    tx, err = db.Begin()
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmGet] ERROR starting read transaction: %s\n", err.Error())
        }
        return nil, -1
    }
    defer tx.Commit()

    kWhere = s


    // build query and fetch fetch results from database

    sqlCheck = fmt.Sprintf("SELECT * FROM %s WHERE %s;", table, kWhere)
    stmt, err = tx.Prepare(sqlCheck)
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmGet] ERROR setting up statement: %s\n", err.Error())
        }
        return nil, -1
    }
    defer stmt.Close()

    // super fugly workaround, till I figured out how to make this dynamic
    if len(keyVs) == 1 {
        rs, err = stmt.Query(keyVs[0])
    } else if len(keyVs) == 2 {
        rs, err = stmt.Query(keyVs[0], keyVs[1])
    } else if len(keyVs) == 3 {
        rs, err = stmt.Query(keyVs[0], keyVs[1], keyVs[2])
    } else if len(keyVs) == 4 {
        rs, err = stmt.Query(keyVs[0], keyVs[1], keyVs[2], keyVs[3])
    } else {
        //err = error( fmt.Sprintf( "Unsupported number of keys given: %d", len(keyVs)) )
    }
    if err != nil {
        if hcIrc.Debugmode {
            fmt.Printf("[DATAMANAGERDEBUG][DmGet] ERROR executing statement: %s\n", err.Error())
        }
        return nil, -1
    }
    defer rs.Close()

    cols, err = rs.Columns()
    numCols = len(cols)
    values = make([]string, numCols)

    i = 0
    for rs.Next() {
        if numCols == 1 {
            err = rs.Scan(&values[0])
        } else if numCols == 2 {
            err = rs.Scan(&values[0], &values[1])
        } else if numCols == 3 {
            err = rs.Scan(&values[0], &values[1], &values[2])
        } else if numCols == 4 {
            err = rs.Scan(&values[0], &values[1], &values[2], &values[3])
        } else if numCols == 5 {
            err = rs.Scan(&values[0], &values[1], &values[2], &values[3], &values[4])
        } else {
            // err
        }

        returnValues = make(map[string]string)
        for j = 0; j < numCols; j++ {
            returnValues[cols[j]] = values[j]
        }
        returnData[i] = returnValues

        i++
    }
    i--

    rs.Close()
    stmt.Close()
    tx.Commit()
    db.Close()

    return returnData, i
}
