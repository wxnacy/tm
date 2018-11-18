package tm

import (
    "testing"
    "strings"
)

func TestFilterStrings(t *testing.T) {
    var texts = []string{
        "deleteFromstring", "deleteFromStringArray", "commandsDeletePreWord",
        "commandsDeleteByBackspace", "resetTables", "resultsSize", "DeleteByString",
    }
    var res []string

    res = FilterStrings(texts, "delete")
    if strings.Join(res, "") != "" {
        t.Error(res, " is error")
    }

    res = FilterStrings(texts, "delStr")
    if strings.Join(res, "") != "" {
        t.Error(res, " is error")
    }
    res = FilterStrings(texts, "table")
    if strings.Join(res, "") != "" {
        t.Error(res, " is error")
    }

    res = FilterStrings([]string{"shop_admin"}, "shoa")
    if strings.Join(res, "") != "" {
        t.Error(res, " is error")
    }

}
