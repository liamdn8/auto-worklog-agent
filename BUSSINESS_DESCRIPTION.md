# User story

I need to create a agent that running on linux ussing golang to support auto development traking. The application should watch the development activity on the IDE include intellij IDE and VS Code

# Assumption

I deployed a ActivityWatch server listening for work event and need to deploy the tracking application by sending API to ActivitiWatch server

The agent runs continuously on a workstation and should auto-discover development workspaces without per-project setup.

# Requirement

1. A work session identified by local git user, project, and branch
2. Jenkins will read from the aw-server to calculate the worklog and automatically put to Jira. So the work session data should support to process easily. The worklog should be automatically by push action or auto detect after idle for 30 mins
3. The tool should easy to install on linux, plug-and-play style app
4. The agent must auto discover the git repository coding. learn how activitywatch work by reading these source code:
    - https://github.com/ActivityWatch/aw-watcher-window/tree/c80aa5adbbe5959fcb661148aeb9f3898e6b68f3
    - https://github.com/ActivityWatch/aw-watcher-input/tree/9bb5045456524b215ae11f422b80ec728c93bac7
    - https://github.com/ActivityWatch/aw-watcher-afk/tree/403a331f6f626afe18094cf61aeed235b75e537c

# Configuration
1. The agent should support configuration file to configure the aw-server endpoint or support argument to override the endpoint
2. The interval to rescan the git repositories is developing is fixed to 5 mins
3. No limit the depth to scan the git repositories (set `maxDepth` to 0 for unlimited; default depth is 5)