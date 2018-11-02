package tm

import (
    "fmt"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
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
        return nil, err
    }

    columns, err := rows.Columns()
    if err != nil {
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

func Insert(db *sql.DB, name string) (int64, error){
    // 准备 sql 语句
    stmt, err := db.Prepare("insert into book (name) values (?)")
    defer stmt.Close()
    if err != nil {
        return 0, err
    }
    // 插入参数并执行语句
    res, err := stmt.Exec(name)
    if err != nil {
        return 0, err
    }
    // 最后插入的 id
    id, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
    return id, nil
}

func InsertTx(db *sql.DB, name string) (int64, error){
    // 开启事务
    tx, err := db.Begin()
    if err != nil {
        return 0, err
    }
    // 准备 sql 语句
    stmt, err := tx.Prepare("insert into book (name) values (?)")
    defer stmt.Close()
    if err != nil {
        return 0, err
    }
    // 插入参数并执行语句
    res, err := stmt.Exec(name)
    if err != nil {
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            return 0, rollbackErr
        }
        return 0, err
    }
    if commitErr := tx.Commit(); commitErr != nil {
        return 0, commitErr
    }
    // 最后插入的 id
    id, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
    return id, nil
}

func Update(db *sql.DB, id int64) error {
    stmt, err := db.Prepare("update book set name = ? where id = ?")
    defer stmt.Close()
    if err != nil {
        return err
    }
    _, err = stmt.Exec("update-name", id)
    if err != nil {
        return err
    }
    return nil
}

func DeleteById(db *sql.DB, id int64) error {
    stmt, err := db.Prepare("delete from book where id = ?")
    defer stmt.Close()
    if err != nil {
        return err
    }
    _, err = stmt.Exec(id)
    if err != nil {
        return err
    }
    return nil
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

