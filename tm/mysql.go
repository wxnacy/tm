package tm

import (
    "fmt"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "strings"
)

const (
    querySqlBegin string = "select show"
    execSqlBegin = "update delete drop create"

    CmdRed = "into values from where order by desc asc index on add table if column null default"
    CmdGreen = "select drop alter insert update delete set explain like and in show create exists not"
    CmdBlue = "count processlist"
)

type Mysql struct {
    db *sql.DB
    user string
    passwd string
    host string
    port string
    database string
}

func NewMysql(user, passwd, host, port, db string) (*Mysql, error) {
    m := &Mysql{
        user:user,
        passwd: passwd,
        host: host,
        port: port,
        database:db,
    }
    m.Connect()
    return m, nil
}

func (this *Mysql) Connect() (error) {
    url := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4",
        this.user, this.passwd, this.host, this.port, this.database,
    )
    DB, err := sql.Open("mysql", url) //第一个参数为驱动名
    if err != nil {
        fmt.Println(err)
        return err
    }
    //设置数据库最大连接数
    DB.SetConnMaxLifetime(100)
    //设置上数据库最大闲置连接数
    DB.SetMaxIdleConns(10)
    //验证连接
    if err := DB.Ping(); err != nil{
        fmt.Println("Connect fail", err.Error())
        return err
    }
    this.db = DB
    return nil
}

func (this *Mysql) Exec(sql string, args ...interface{}) (sql.Result, error) {
    stmt, err := this.db.Prepare(sql)
    if err != nil {
        return nil, err
    }
    res, err := stmt.Exec(args...)
    defer stmt.Close()
    return res, err
}

func (this *Mysql) Query(
    query string, args ...interface{},
) ([]map[string]string, error){
    rows, err := this.db.Query(query, args...)
    defer rows.Close()
    checkErr(err)

    columns, err := rows.Columns()
    checkErr(err)
    values := make([][]byte, len(columns))
    scans := make([]interface{}, len(columns))
    //让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

    res := make([]map[string]string, 0)

    for rows.Next() {
        var item = make(map[string]string)
        rows.Scan(scans...)

        for i, d := range values {
            item[columns[i]] = string(d)
        }

        res = append(res, item)
    }

    return res, nil
}

func (this *Mysql) QueryResultArray(
    query string, args ...interface{},
) ([][]string, error){
    rows, err := this.db.Query(query, args...)
    if err != nil {
        Log.Error(err)
        return nil, err
    }

    columns, err := rows.Columns()
    if err != nil {
        Log.Error(err)
        return nil, err
    }
    values := make([][]byte, len(columns))
    scans := make([]interface{}, len(columns))
    //让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

    res := make([][]string, 0)
    res = append(res, columns)

    for rows.Next() {
        var item = make([]string, 0)
        rows.Scan(scans...)

        for _, d := range values {
            item = append(item, string(d))
        }
        res = append(res, item)
    }

    defer rows.Close()
    return res, nil
}

func (this *Mysql) Close() error{
    return this.db.Close()
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func IsQuerySql(cmd string) (flag bool) {
    flag = false
    if cmd == "" {
        return
    }
    cmdBegin := strings.ToLower(strings.Split(cmd, " ")[0])
    i := inArray(cmdBegin, strings.Split(querySqlBegin, " "))
    if i > -1 {
        flag = true
    }
    return
}

func isShowTablesFrames(cmd string, index int) (flag bool) {
    flag = false
    if index < 2 {
        return
    }
    if len(cmd) < index {
        return
    }
    preRune := []rune(cmd)[index-1]
    prePreRune := []rune(cmd)[index-2]

    preWord := stringPreWord(cmd, index)
    flag1 := preWord == "from" || preWord == "update" || preWord == "table"
    flag2 := preRune == ' ' && prePreRune != ' '
    flag = flag1 && flag2
    return
}

func isHideTablesFrames(cmd string, index int) (flag bool) {
    flag = true
    if index < 2 {
        return
    }
    if len(cmd) < index {
        return
    }
    preRune := []rune(cmd)[index-1]
    if preRune == 0 {
        return
    }
    // prePreRune := []rune(cmd)[index-2]

    preWord := stringPreWord(cmd, index)
    flag1 := preWord != "from" && preWord != "update" && preWord != "table"
    flag2 := preRune == ' '
    // flag2 := preRune == ' ' && prePreRune != ' '
    if flag1 && flag2 {
        return
    }

    if preWord == "from" && preRune == 'm' {
        return
    }

    if preWord == "update" && preRune == 'e' {
        return
    }
    if preWord == "table" && preRune == 'e' {
        return
    }

    flag = false
    return
}
