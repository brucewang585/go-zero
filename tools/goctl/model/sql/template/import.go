package template

var (
	Imports = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/brucewang585/go-zero/core/stores/cache"
	"github.com/brucewang585/go-zero/core/stores/sqlc"
	"github.com/brucewang585/go-zero/core/stores/sqlx"
	"github.com/brucewang585/go-zero/core/stringx"
	"github.com/brucewang585/go-zero/tools/goctl/model/sql/builderx"
)
`
	ImportsNoCache = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/brucewang585/go-zero/core/stores/sqlc"
	"github.com/brucewang585/go-zero/core/stores/sqlx"
	"github.com/brucewang585/go-zero/core/stringx"
	"github.com/brucewang585/go-zero/tools/goctl/model/sql/builderx"
)
`
)
