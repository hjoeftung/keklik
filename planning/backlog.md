# Keklik Backend Backlog

This document decomposes [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md) into S-sized implementation tasks.

## S Size Definition

- Size: `S`
- Expected effort: about 0.5 to 2 developer days
- A task is complete only when code, tests, and minimal documentation are included

## Recommended Delivery Order

1. Foundation and architecture
2. Authentication and family onboarding
3. Sleep lifecycle
4. Reporting and summaries
5. Operational hardening
6. Deployment and early operations

## Story Files

- [planning/backlog/foundation.md](/home/hjoeftung/code/projects/keklik/planning/backlog/foundation.md)
- [planning/backlog/family.md](/home/hjoeftung/code/projects/keklik/planning/backlog/family.md)
- [planning/backlog/sleep.md](/home/hjoeftung/code/projects/keklik/planning/backlog/sleep.md)
- [planning/backlog/summaries.md](/home/hjoeftung/code/projects/keklik/planning/backlog/summaries.md)
- [planning/backlog/operations.md](/home/hjoeftung/code/projects/keklik/planning/backlog/operations.md)
- [planning/backlog/deployment.md](/home/hjoeftung/code/projects/keklik/planning/backlog/deployment.md)


## Suggested First Milestone

If you want an initial deliverable that already has end-user value for a simple client, stop after:
- [TASK-001](#task-001-bootstrap-go-service-skeleton)
- [TASK-002A](#task-002a-dockerize-the-app-and-bootstrap-docker-compose-with-postgresql)
- [TASK-003](#task-003-add-postgresql-migration-framework-and-baseline-schema)
- [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
- [TASK-006](#task-006-implement-create-family-use-case-and-api)
- [TASK-007](#task-007-implement-google-oauth-identity-flow)
- [TASK-008](#task-008-implement-edit-family-use-case-and-api)
- [TASK-011](#task-011-model-sleep-session-aggregate-and-repository-interfaces)
- [TASK-012](#task-012-implement-timezone-aware-sleep-classification-service)
- [TASK-013](#task-013-implement-start-sleep-use-case-and-api)
- [TASK-014](#task-014-implement-stop-sleep-use-case-and-api)
- [TASK-017](#task-017-implement-sleep-history-query-and-api)

That milestone would support:
- authentication
- family creation
- family settings
- sleep start and stop
- basic sleep history

## Notes for Planning

- Keep tasks small. If any task starts expanding beyond roughly 2 days, split by transport layer, application service, or repository work.
- Prefer finishing domain model and tests before wiring HTTP handlers.
- Treat timezone and daylight saving tests as mandatory, not optional cleanup.
- Preserve the stored classification result on historical sleep sessions so later family-setting changes do not rewrite history implicitly.