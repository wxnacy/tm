package main

import (
    "github.com/wxnacy/tm/tm"
    "fmt"
    "time"
    "flag"
    "os"
    "strings"
    "database/sql"
	"github.com/c-bata/go-prompt"
    "github.com/howeyc/gopass"
    "os/user"
    "github.com/olekukonko/tablewriter"
)

const (
    version string = "0.4.0"
)

var m *tm.Mysql
var t *tm.Terminal
var C tm.Credential
var err error
var args []string
var conf string
var username string
var passwd string
var pass []byte
var host string
var port string
var db string
var dbname string
var creDir = os.Getenv("HOME") + "/.tm/credentials"

var v bool
var p bool
var s string
var c string

func InitArgs() {
    flag.BoolVar(&v, "v", false, "")
    flag.StringVar(&username, "u", "", "")
    flag.StringVar(&host, "h", "", "")
    flag.StringVar(&port, "P", "", "")
    flag.StringVar(&s, "s", "", "")
    flag.StringVar(&c, "c", "", "")
    flag.BoolVar(&p, "p", false, "")
    flag.Parse()
    args = flag.Args()

    if c != "" {
        path := fmt.Sprintf("%s/%s", creDir, c)
        C, err = tm.LoadCredentialFromPath(path)
        PrintErr(err)
        InitMysql()

    }
    if len(args) > 0 {
        dbname = args[0]
    }
    if p {
        fmt.Print("Enter password: ")
        pass, err = gopass.GetPasswd()
        checkErr(err)
        passwd = string(pass)
    }
    conf = ""
    C = tm.Credential{}
    C.Username = username
    C.Host = host
    C.Port = port
    C.Password = passwd
    C.Database = dbname

    if s != "" {
        tm.SaveCredential(s, C)
    }

    tm.Log.Info("args ", username, host, port, passwd, s, dbname)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

// func SaveMysqlConfig() {
    // var credentialsDir = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".tm/credentials")
    // filename := credentialsDir + "/" + conf
    // tm.SaveFile(filename, fmt.Sprintf("%s %s %s %s %s", username, passwd, host, port, db))
// }

func InitMysql() {

    username = C.Username
    if username == "" {
        osUser, err := user.Current()
        PrintErr(err)
        username = osUser.Username
    }

    passwd = C.Password
    host = C.Host
    if host == "" {
        host = "localhost"
    }

    port = C.Port
    if port == "" {
        port = "3306"
    }

    dbname = C.Database

    m, err = tm.NewMysql(username, passwd, host, port, dbname)
    PrintErr(err)

    if dbname != "" {
        InitTerminal()
    }
    res, err := m.QueryResultArray("show databases;")
    PrintErr(err)

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{res[0][0]})

    for _, d := range res[1:] {
        table.Append([]string{d[0]})
    }
    table.Render() // Send output
    fmt.Println("Input command: [use <database>;] to select database")
    // fmt.Println(`
    // Input command
        // use <database>
    // To select databse
    // `)
    Prompt()

}

func PrintErr(err error) {
    tm.Log.Error(err)
    if err != nil {
        fmt.Println(err)
        os.Exit(0)
        return
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

func completer(d prompt.Document) []prompt.Suggest {
    s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by username"},
		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(t string) {
    tm.Log.Info("input ", t)
    t = strings.TrimRight(t, ";")
    ts := strings.Split(t, " ")
    cmd := ts[0]
    switch cmd {
        case "exit": {
            tm.Log.Info("Exit")
            os.Exit(0)
        }
        case "": {
            tm.Log.Info(username)
            if username == "" {
                username = "root"
                LivePrefixState.IsEnable = true
                LivePrefixState.LivePrefix = cmd
            }
        }
        case "use": {
            dbname = ts[1]
            m.SetDatabase(dbname)
            InitTerminal()

        }
        default: {
            if username == "" {
                username = cmd
            }

            LivePrefixState.LivePrefix = cmd + "> "
            LivePrefixState.IsEnable = true
        }
    }
    fmt.Println(username)
	return
}

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}

func Prompt() {

    p := prompt.New(
        executor,
        completer,
        prompt.OptionPrefix("tm > "),
    )
    p.Run()
}

func InitTerminal() {

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

func main() {

    InitArgs()
    if v {
        fmt.Println(version)
        return
    }
    InitMysql()

}


