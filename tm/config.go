package tm

import (
    "gopkg.in/ini.v1"
    "fmt"
)

var err error

type Credential struct {
    Username string `ini:"user"`
    Host string `ini:"host"`
    Password string `ini:"password"`
    Port string `ini:"port"`
    Database string `ini:"database"`
}


func LoadCredentialFromPath(path string) (c Credential, err error){
    c = Credential{}

    err = ini.MapTo(&c, path)
    if err != nil {
        return c, err
    }
    return c, nil
}

func SaveCredential( credentialName string, c Credential) {
    path := fmt.Sprintf("%s/%s", CREDENTIAL_DIR, credentialName)
    flag := IsFile(path)
    if !flag {
        SaveFile(path, "")
    }
    cfg, err := ini.Load(path)
    PrintErr(err)
    if c.Username != "" {
        cfg.Section("").Key("user").SetValue(c.Username)
    }
    if c.Password != "" {
        cfg.Section("").Key("password").SetValue(c.Password)
    }
    if c.Host != "" {
        cfg.Section("").Key("host").SetValue(c.Host)
    }
    if c.Port != "" {
        cfg.Section("").Key("port").SetValue(c.Port)
    }
    if c.Database != "" {
        cfg.Section("").Key("database").SetValue(c.Database)
    }
    cfg.SaveTo(path)
}
