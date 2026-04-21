package data

// AnyItem is the discriminated union of all item kinds.
type AnyItem interface {
	ItemKind() string
	ItemID() string
	ItemName() string
}

type ServiceStatus struct {
	Active       string // "running" | "stopped" | "failed" | "unknown"
	Since        string
	SocketHealth string // "ok" | "stale" | "unknown"
	Version      string
}

type AppData struct {
	Skills    []OcSkill
	Hooks     []OcHook
	Models    []OcModel
	Workspace []OcWorkspaceFile
	MCP       []OcMcpServer
	Sessions  []OcSession
	Cron      []OcCronJob
	Memory    []OcMemoryChunk
	Updates   []OcUpdateRelease
	Webhooks  []OcWebhook
	AuditLog  []OcAuditEntry
}

// ─── Concrete item types ─────────────────────────────────────────────────────

type OcSkill struct {
	ID          string
	Name_       string
	Description string
	Scope       string // "built-in" | "installed"
	FilePath    string
	FullContent string
}

func (s OcSkill) ItemKind() string { return "skill" }
func (s OcSkill) ItemID() string   { return s.ID }
func (s OcSkill) ItemName() string { return s.Name_ }

type OcHook struct {
	ID          string
	Name_       string
	Description string
	Enabled     bool
	RawConfig   map[string]any
}

func (h OcHook) ItemKind() string { return "hook" }
func (h OcHook) ItemID() string   { return h.ID }
func (h OcHook) ItemName() string { return h.Name_ }

type OcModel struct {
	ID            string
	Name_         string
	Provider      string
	Reasoning     bool
	ContextWindow int
	MaxTokens     int
	CostInput     float64
	CostOutput    float64
}

func (m OcModel) ItemKind() string { return "model" }
func (m OcModel) ItemID() string   { return m.ID }
func (m OcModel) ItemName() string { return m.Name_ }

type OcWorkspaceFile struct {
	ID           string
	Name_        string
	FilePath     string
	Preview      string
	FullContent  string
	WordCount    int
	LastModified string
}

func (w OcWorkspaceFile) ItemKind() string { return "workspace" }
func (w OcWorkspaceFile) ItemID() string   { return w.ID }
func (w OcWorkspaceFile) ItemName() string { return w.Name_ }

type McpDependency struct {
	Name string
	Met  bool
}

type OcMcpServer struct {
	ID           string
	Name_        string
	URL          string
	Command      string
	Args         []string
	Transport    string
	Headers      map[string]string
	Enabled      bool
	Available    bool
	Dependencies []McpDependency
}

func (m OcMcpServer) ItemKind() string { return "mcp" }
func (m OcMcpServer) ItemID() string   { return m.ID }
func (m OcMcpServer) ItemName() string { return m.Name_ }

type OcSession struct {
	ID          string
	Name_       string
	Channel     string
	UpdatedAt   int64 // ms timestamp
	SessionFile string
	SizeKb      int
}

func (s OcSession) ItemKind() string { return "session" }
func (s OcSession) ItemID() string   { return s.ID }
func (s OcSession) ItemName() string { return s.Name_ }

type OcCronJob struct {
	ID          string
	Name_       string
	Schedule    string
	Command     string
	Enabled     bool
	Description string
}

func (c OcCronJob) ItemKind() string { return "cron" }
func (c OcCronJob) ItemID() string   { return c.ID }
func (c OcCronJob) ItemName() string { return c.Name_ }

type OcMemoryChunk struct {
	ID        string
	Name_     string
	Path      string
	Source    string
	StartLine int
	EndLine   int
	Model     string
	Text      string
	UpdatedAt int64
}

func (m OcMemoryChunk) ItemKind() string { return "memory" }
func (m OcMemoryChunk) ItemID() string   { return m.ID }
func (m OcMemoryChunk) ItemName() string { return m.Name_ }

type InstallRecord struct {
	From      string
	To        string
	Timestamp string
}

type OcUpdateRelease struct {
	ID           string
	Name_        string
	Version      string
	IsInstalled  bool
	IsLatest     bool
	IsAvailable  bool
	LastChecked  string
	ChangeCount  int
	Changes      []string
	Fixes        []string
	InstallRecord *InstallRecord
}

func (u OcUpdateRelease) ItemKind() string { return "update" }
func (u OcUpdateRelease) ItemID() string   { return u.ID }
func (u OcUpdateRelease) ItemName() string { return u.Name_ }

type OcWebhook struct {
	ID           string
	Name_        string
	Enabled      bool
	Path         string
	SessionKey   string
	Secret       string
	ControllerID string
	Description  string
}

func (w OcWebhook) ItemKind() string { return "webhook" }
func (w OcWebhook) ItemID() string   { return w.ID }
func (w OcWebhook) ItemName() string { return w.Name_ }

type OcAuditEntry struct {
	ID                string
	Name_             string
	TS                string
	Event             string
	Source            string
	ConfigPath        string
	Command           string
	Argv              []string
	PID               int
	Result            string
	Suspicious        []string
	PreviousBytes     *int64
	NextBytes         int64
	PreviousHash      string
	NextHash          string
	GatewayModeBefore *string
	GatewayModeAfter  *string
}

func (a OcAuditEntry) ItemKind() string { return "auditlog" }
func (a OcAuditEntry) ItemID() string   { return a.ID }
func (a OcAuditEntry) ItemName() string { return a.Name_ }

// ─── Transcript types ─────────────────────────────────────────────────────────

type TurnKind string

const (
	TurnText       TurnKind = "text"
	TurnToolCall   TurnKind = "tool_call"
	TurnToolResult TurnKind = "tool_result"
)

type Turn struct {
	Kind      TurnKind
	Role      string
	Text      string
	ID        string
	ToolName  string
	Input     string
	ToolUseID string
	Content   string
}
