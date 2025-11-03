# User story

I need to create a agent that running on linux ussing golang to support auto development traking. The application should watch the development activity on the IDE include intellij IDE and VS Code

# Assumption

I deployed a ActivityWatch server listening for work event and need to deploy the tracking application by sending API to ActivitiWatch server

The agent runs continuously on a workstation and should auto-discover development workspaces without per-project setup.

# Requirement

1. A work session identified by local git user, project, and branch
2. Jenkins will read from the aw-server to calculate the worklog and automatically put to Jira. So the work session data should support to process easily. The worklog should be automatically by push action or auto detect after idle for 30 mins
3. The tool should easy to install on linux, plug-and-play style app
4. The agent must discover IDE session which editing a git repositories globally (configurable roots) so IDE activity for any project on the machine is captured automatically. By default the agent scans `/home/${USER}` to a depth of 5 levels to find repositories.

# Configuration
1. The agent should support configuration file to configure the aw-server endpoint or support argument to override the endpoint
2. The interval to rescan the git repositories is developing is fixed to 5 mins
3. No limit the depth to scan the git repositories (set `maxDepth` to 0 for unlimited; default depth is 5)