# User story

I need to create a agent that running on linux ussing golang to support auto development traking. The application should watch the development activity on the IDE include intellij IDE and VS Code

# Assumption

I deployed a ActivityWatch server listening for work event and need to deploy the tracking application by sending API to ActivitiWatch server

# Requirement

1. A work session identified by local git user, project, and branch
2. Jenkins will read from the aw-server to calculate the worklog and automatically put to Jira. So the work session data should support to process easily. The worklog should be automatically by push action or auto detect after idle for 30 mins
3. The tool should easy to install on linux, plug-and-play style app