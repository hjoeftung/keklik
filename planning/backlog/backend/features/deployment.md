## Story 6: Deployment and Early Operations

### TASK-025: Define the single-VM production runtime contract
- Size: `S`
- Goal: Make the backend deployable on one VM without introducing environment-specific guesswork.
- Scope:
  - Define the production container topology for one VM
  - Decide the initial public base URL as `https://mioqo.info`
  - Document required runtime configuration, secrets, host ports, and persistent volumes
  - Define how the application image tag is selected for deployment
- Proposed topology:
  - One reverse proxy container for public ingress and TLS
  - One API container built from the published application image
  - One PostgreSQL container with a named volume on the VM
- Suggested implementation details:
  - Keep Compose files production-oriented rather than reusing local-development defaults blindly
  - Store non-secret defaults in versioned config examples
  - Keep secrets out of Git and inject them at deploy time
- Dependencies:
  - [TASK-002](#task-002-add-configuration-and-environment-contract)
  - [TASK-002A](#task-002a-dockerize-the-app-and-bootstrap-docker-compose-with-postgresql)
  - [TASK-024](#task-024-add-github-actions-ci-and-ghcr-image-publishing)
- Acceptance criteria:
  - The repository documents the exact containers, ports, volumes, and environment variables needed for production
  - The public base URL is defined consistently as `https://mioqo.info`
  - A developer can determine which image tag should be deployed without manual interpretation

### TASK-026: Add a production Compose stack with HTTPS ingress for mioqo.info
- Size: `S`
- Goal: Run the backend on the VM behind a minimal reverse proxy with automatic TLS.
- Scope:
  - Add a production-oriented Compose file for the VM
  - Add a reverse proxy container that terminates TLS and forwards traffic to the API container
  - Persist reverse-proxy certificate state and PostgreSQL data on the VM
  - Expose only the ports required for HTTP and HTTPS from the host
- Recommended approach:
  - Use Caddy because it keeps automatic HTTPS and renewal simpler than Nginx plus Certbot for a single-host setup
  - Route `mioqo.info` to the API service and redirect plain HTTP to HTTPS
  - Keep the API container off the public host network except through the reverse proxy
- Dependencies:
  - [TASK-025](#task-025-define-the-single-vm-production-runtime-contract)
- Acceptance criteria:
  - Bringing up the production Compose stack serves the API through `https://mioqo.info`
  - HTTP requests are redirected to HTTPS
  - Reverse-proxy certificate data survives container restarts
  - The API remains reachable internally from the proxy and is not exposed directly on a public high port

### TASK-027: Automate VM provisioning and deployment with Ansible
- Size: `S`
- Goal: Make provisioning and deployment repeatable enough to use from the start of development.
- Scope:
  - Add an Ansible inventory and role or playbook structure for the single VM
  - Install Docker Engine and the Docker Compose plugin on a clean VM
  - Create the application directory structure, upload Compose and config assets, and manage secrets injection
  - Pull the selected container image and apply the Compose stack idempotently
  - Run or trigger database migrations as part of deployment if the app requires them
- Suggested implementation details:
  - Keep one playbook for host bootstrap and one for app deployment
  - Parameterize the deployed image tag so trunk deployments can use a specific published SHA image
  - Use Ansible Vault or an external secret source for production secrets
- Dependencies:
  - [TASK-026](#task-026-add-a-production-compose-stack-with-https-ingress-for-mioqoinfo)
- Acceptance criteria:
  - A clean VM can be prepared with a documented Ansible command
  - Re-running the deployment playbook is idempotent and does not require manual host edits
  - Production secrets are not committed to the repository in plain text
  - The deployment process can target a specific published image tag

### TASK-028: Define the trunk-based deployment flow, smoke checks, and rollback path
- Size: `S`
- Goal: Turn every merge to trunk into a deployable artifact with a simple, explicit release procedure.
- Scope:
  - Define trunk as the only long-lived branch for releasable changes
  - Document how merges produce deployable images through the existing publish pipeline
  - Add a deployment runbook that covers deploy, verify, and rollback on the VM
  - Add a minimal smoke-check step that verifies the public endpoint after deployment
- Recommended operating model:
  - Every merge to trunk publishes an immutable image tag
  - Deployment to the VM is triggered manually at first by running the Ansible deploy playbook with that image tag
  - Rollback means redeploying the previous known-good image tag
- Dependencies:
  - [TASK-027](#task-027-automate-vm-provisioning-and-deployment-with-ansible)
- Acceptance criteria:
  - The team has one documented command path for deploying a chosen image tag to the VM
  - Post-deploy verification includes a public smoke check against `https://mioqo.info/healthz`
  - Rollback steps are documented and rely on redeploying a prior image tag rather than mutating the server manually
  - The release procedure is simple enough to use during early feature development without creating a separate release branch