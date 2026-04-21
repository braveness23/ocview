package data

func LoadAll() AppData {
	return AppData{
		Skills:    LoadSkills(),
		Hooks:     LoadHooks(),
		Models:    LoadModels(),
		Workspace: LoadWorkspace(),
		MCP:       LoadMCP(),
		Sessions:  LoadSessions(),
		Cron:      LoadCron(),
		Memory:    LoadMemory(),
		Updates:   LoadUpdates(),
		Webhooks:  LoadWebhooks(),
		AuditLog:  LoadAuditLog(),
	}
}
