package main

import (
    "github.com/wxnacy/tm/tm"
    "fmt"
    "time"
    "flag"
    "os"
    "strings"
    "database/sql"
)

const (
    version string = "0.3.3"
)

var m *tm.Mysql
var t *tm.Terminal
var err error
var args []string
var conf string
var user string
var passwd string
var host string
var port string
var db string
var creDir = os.Getenv("HOME") + "/.tm/credentials"

var v bool

func InitArgs() {
    flag.BoolVar(&v, "v", false, "")
    flag.Parse()
    args = flag.Args()
    conf = ""
    user = "root"
    host = "localhost"
    port = "3306"
}

func InitMysqlConfig() {

    fmt.Print("user(root): ")
    fmt.Scanln(&user)
    fmt.Print("password: ")
    fmt.Scanln(&passwd)
    fmt.Print("host(localhost): ")
    fmt.Scanln(&host)
    fmt.Print("port(3306): ")
    fmt.Scanln(&port)
    fmt.Print("db: ")
    fmt.Scanln(&db)

}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func SaveMysqlConfig() {
    var credentialsDir = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".tm/credentials")
    filename := credentialsDir + "/" + conf
    tm.SaveFile(filename, fmt.Sprintf("%s %s %s %s %s", user, passwd, host, port, db))
}

func InitMysql() {

    if len(args) == 0 {
        fmt.Println("Connect to Mysql")
        InitMysqlConfig()
    } else {
        var action = args[0]
        if action == "init" {
            fmt.Println("Save to Mysql")
            fmt.Print("Config name: ")
            fmt.Scanln(&conf)
            InitMysqlConfig()
            SaveMysqlConfig()
        } else {
            conf = action
            confData, err := tm.ReadFile(creDir + "/" + action)
            checkErr(err)
            urls := strings.Split(confData, " ")
            // fmt.Println(urls)
            user = urls[0]
            passwd = urls[1]
            host = urls[2]
            port = urls[3]
            db = urls[4]
        }
    }

    m, err = tm.NewMysql(user, passwd, host, port, db)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
}

func QueryTables() []string{
    res, err := m.QueryResultArray("show tables")
    if err != nil {
        panic(err)
    }

    ts := make([]string, 0)
    for _, d := range res {
        ts = append(ts, d[0])
    }
    return ts
}

func onExecCommands(cmds []string) {

    begin := time.Now()
    cmd := cmds[0]

    var results [][]string
    var err error
    var rowsAffected int64
    var res sql.Result
    if tm.IsQuerySql(cmd) {
        results, err = m.QueryResultArray(cmd)
        rowsAffected = int64(len(results) - 1)
    } else if strings.HasPrefix(cmd, "-- ") {

    } else {
        res, err = m.Exec(cmd)
        if err == nil {
            rowsAffected, err = res.RowsAffected()
        }
    }

    if err != nil {
        t.SetResultsBottomContent(err.Error())
        t.SetResultsIsError(true)
        t.ClearResults()
    } else {
        t.SetResults(results)
        t.SetResultsIsError(false)
        c := fmt.Sprintf(
            "No Erros; %d rows affected, taking %v",
            rowsAffected,
            time.Since(begin),
        )
        t.SetResultsBottomContent(c)
    }

}

func onReload(typ tm.ReloadType, v ...interface{}) {
    begin := time.Now()
    if typ == tm.ReloadTypeAllTable {
        tables := QueryTables()
        t.SetTables(tables)

        var tablesFields = make(map[string][]string, 0)

        for _, d := range tables[1:] {

            res, _ := m.QueryResultArray(fmt.Sprintf("select * from `%s` limit 1;", d))
            var fields []string
            fields = res[0]
            tablesFields[d] = fields
            tm.Log.Infof("tablesFields %v", fields)
        }

        t.SetTablesFields(tablesFields)
        c := fmt.Sprintf(
            "No Erros; %d rows affected, taking %v",
            len(tables) - 1,
            time.Since(begin),
        )
        t.SetResultsBottomContent(c)

    }
}

func main() {
    InitArgs()
    if v {
        fmt.Println(version)
        return
    }
    InitMysql()

    t, err = tm.New(conf)
    if err != nil {
        tm.Log.Error("New Terminal ", err)
        panic(err)
    }

    tables := QueryTables()
    t.SetTables(tables)
    t.OnExecCommands(onExecCommands)
    t.OnReload(onReload)

    for {
        t.Rendering()
        t.ListenKeyBorad()
    }

    defer m.Close()
    defer t.Close()
}

