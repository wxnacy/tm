package tm

import (
    "testing"
    "fmt"
)

func TestLoadCredentialFromPath(t *testing.T) {
    c, err := LoadCredentialFromPath("/Users/wxnacy/.tm/credentials/test")
    if err != nil {
        t.Error(err)
    }
}
