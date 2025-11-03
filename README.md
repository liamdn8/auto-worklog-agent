# Auto Worklog Agent

Go-based agent that observes development activity within configured Git repositories, converts active editing windows into work sessions, and publishes them to an ActivityWatch server. Sessions are keyed by Git user, project, and branch to support downstream automation (e.g., Jenkins aggregation and Jira worklog updates).

## Features
- Watches repository trees for change events (VS Code, IntelliJ, or other tools writing to disk).
- Automatically closes sessions after 30 minutes of inactivity (configurable).
- Publishes aggregated work-session events to ActivityWatch buckets for easy downstream processing.
- Simple CLI with JSON configuration.

## Prerequisites
- Go 1.21+
- Docker (for running a local ActivityWatch server)
- Git repositories available on the local filesystem

## Quick Start

1. **Start ActivityWatch server (test setup):**

	 ```bash
	 docker compose up -d
	 ```

	 This launches `aw-server` at `http://localhost:5600` with data persisted in the `aw_data` volume.

2. **Create a configuration file (optional when defaults suffice):**

	 ```jsonc
	 {
		 "activityWatch": {
			 "baseURL": "http://localhost:5600",
			 "bucketPrefix": "awagent",
			 "machine": "developer-workstation"
		 },
		 "git": {
			 "repositories": [
				 "/home/dev/projects/sample-app"
			 ]
		 },
		 "session": {
			 "idleTimeoutMinutes": 30,
			 "pollInterval": "5s",
			 "flushInterval": "15s"
		 }
	 }
	 ```

	 Save the file as `config.json` (or another path of your choice).

3. **Run the agent:**

	 ```bash
	 go run ./cmd/awagent --config ./config.json
	 ```

	 Without `--config`, the agent watches the current working directory and uses default timing values.

## How It Works
- `fsnotify` monitors each configured repository recursively, skipping heavy directories (`node_modules`, `vendor`, `target`, `build`).
- Activity bumps extend the active session; inactivity beyond the configured timeout flushes the session to ActivityWatch.
- Events include Git metadata (user, email, remote, branch) making it straightforward for Jenkins to calculate worklogs and sync with Jira.

## Next Steps
- Add IDE-specific integrations (e.g., richer metadata from VS Code or IntelliJ APIs).
- Implement push detection hooks to flush sessions immediately after `git push` operations.
- Provide Jenkins pipeline examples that consume the ActivityWatch session data.

## Development

```bash
go test ./...
```

To rebuild binaries:

```bash
go build ./cmd/awagent
```
