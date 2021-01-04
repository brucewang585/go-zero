package template

var Error = `package {{.pkg}}

import "github.com/brucewang585/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound
`
