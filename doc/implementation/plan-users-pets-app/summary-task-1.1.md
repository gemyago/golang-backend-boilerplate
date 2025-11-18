# Task 1.1: Define OpenAPI specification for users and pets endpoints

## Summary

Updated `internal/api/http/v1routes.yaml` to include the complete OpenAPI specification for users and pets endpoints as defined in the plan.

### Changes Made

- Updated OpenAPI version from "3.0.0" to "3.0.3"
- Updated API title from "Minimalistic openapi starter" to "Users & Pets API"
- Removed unnecessary description and license fields to match the plan specification
- Verified that all users and pets endpoints were already present in the file:
  - Users endpoints: POST/GET /users, GET/PUT/DELETE /users/{userId}
  - Pets endpoints: GET/POST /users/{userId}/pets, DELETE /users/{userId}/pets/{petId}
- Verified all request/response schemas were properly defined

### Verification

- **Lint**: `make lint` passes with 0 issues
- **YAML Validation**: OpenAPI specification is valid and follows the plan requirements
- **Schema Completeness**: All required schemas, responses, and error handling are defined

The OpenAPI specification now fully matches the requirements from `doc/plan-users-pets-app.md` and is ready for code generation in the next task.