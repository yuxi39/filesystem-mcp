package root

import (
	"time"

	"github.com/yuxi39/filesystem-mcp/internal/innerpath"
)

type RootStatus string

const (
	RootStatusUnknown   RootStatus = "unknown"
	RootStatusReachable RootStatus = "reachable"
	RootStatusMissing   RootStatus = "missing"
	RootStatusDenied    RootStatus = "permission_denied"
)

type Root struct {
	PathType     innerpath.PathKind `json:"pathType" jsonschema:"Normalized path kind used internally, such as win_drive, win_unc, or unix_abs."`
	NameSpace    string             `json:"namespace" jsonschema:"Unique root namespace used before ':' in namespace paths, for example odds:README.md."`
	InnerPath    string             `json:"innerPath" jsonschema:"Platform-independent internal path represented as slash-joined normalized segments."`
	AbsolutePath string             `json:"absolutePath" jsonschema:"Native absolute filesystem path passed to os and filepath operations."`
	PathSegments []string           `json:"pathSegments" jsonschema:"Normalized platform-independent path segments used by the root tree."`
	Status       RootStatus         `json:"status" jsonschema:"Last known reachability status of this root."`
	LastChecked  time.Time          `json:"lastChecked,omitempty" jsonschema:"Last time the root path was checked for reachability."`
	LastError    string             `json:"lastError,omitempty" jsonschema:"Last reachability error, if any."`
}
