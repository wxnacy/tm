package main

import (
    "github.com/wxnacy/tm/tm"
    "fmt"
    "time"
    "flag"
    "os"
    "strings"
)

const (
    version string = "0.1.3"
)

var m *tm.Mysql
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
            fmt.Println(urls)
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

func OpenTable(name string) [][]string {

    sql := fmt.Sprintf("select * from %s limit 10", name)
    results, err := m.QueryResultArray(sql)
    if err != nil {
        panic(err)
    }
    return results
}

func main() {
    InitArgs()
    if v {
        fmt.Println(version)
        return
    }
    InitMysql()

    t, err := tm.New(conf)
    if err != nil {
        panic(err)
    }

    tables := QueryTables()
    t.SetTables(tables)
    if len(tables) >= 2 {
        t.SetResults(OpenTable(tables[1]))
    }
    // t.SetResults(OpenTable("ad"))
    t.OnExecCommands(func (cmds []string) {
        // time.Sleep(1 * time.Second)
        begin := time.Now()
        sql := fmt.Sprintf(cmds[0])
        results, err := m.QueryResultArray(sql)
        end := time.Now()

        dur := end.Sub(begin).Nanoseconds()
        if err != nil {
            t.SetResultsBottomContent(err.Error())
            t.SetResultsIsError(true)
            t.ClearResults()
        } else {
            t.SetResults(results)
            t.SetResultsIsError(false)
            c := fmt.Sprintf("No Erros; taking %d ms", dur/10000)
            t.SetResultsBottomContent(c)
        }

    })

    for {
        // if t.IsListenKeyBorad() {
        t.Rendering()
        t.ListenKeyBorad()
        // }
    }

    defer m.Close()

    defer t.Close()
}

