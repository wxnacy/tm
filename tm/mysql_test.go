package tm

import (
    "testing"
)

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

    flag = isShowTablesFrames("from shop, ", 6)
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

func TestSqlKeyWordIndexs(t *testing.T) {

    res := sqlKeyWordIndexs("select * from shop where ")
    rightFlag := res["select"] == 0 && res["from"] == 9 && res["where"] == 19
    if !rightFlag {
        t.Error(res, " is error")
    }

    res = sqlKeyWordIndexs("update shop set where ")
    rightFlag = res["update"] == 0 && res["set"] == 12 && res["where"] == 16
    if !rightFlag {
        t.Error(res, " is error")
    }

    res = sqlKeyWordIndexs("delete from shop where ")
    rightFlag = res["delete"] == 0 && res["from"] == 7 && res["where"] == 17
    if !rightFlag {
        t.Error(res, " is error")
    }
}
