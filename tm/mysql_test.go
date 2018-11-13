package tm

import (
    "testing"
)


// func TestQueryResultArray(t *testing.T) {

    // m, err := NewMysql("root", )
    // if err != nil {
        // panic(err)
    // }
    // defer m.Close()
    // results, err := m.QueryResultArray("select * from user")
    // if err != nil {
        // t.Error(err)
    // } else {
        // t.Log(results)
    // }

// }

func TestIsQuerySql(t *testing.T) {
    var flag bool
    flag = IsQuerySql("select * ")
    if !flag {
        t.Error(flag, " is error")
    }
    flag = IsQuerySql("update * ")
    if flag {
        t.Error(flag, " is error")
    }
    flag = IsQuerySql("")
    if flag {
        t.Error(flag, " is error")
    }
    flag = IsQuerySql("select")
    if !flag {
        t.Error(flag, " is error")
    }
}
func TestIsShowTablesFrames(t *testing.T) {

    var flag bool
    flag = isShowTablesFrames("from ", 4)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("from ", 5)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("from  ", 6)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("from  ", 7)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("from  ", 1)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("from  ", 3)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("fr ", 1)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("update ", 7)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isShowTablesFrames("table ", 6)
    if !flag {
        t.Error(flag, " is error")
    }
}

func TestIsHideTablesFrames(t *testing.T) {

    var flag bool
    flag = isHideTablesFrames("from ", 4)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("from ", 1)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("from ", 7)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("from ", 5)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("from a", 6)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("update ", 6)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("update ", 7)
    if flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("table ", 5)
    if !flag {
        t.Error(flag, " is error")
    }
    flag = isHideTablesFrames("table ", 6)
    if flag {
        t.Error(flag, " is error")
    }
}
