# Design Spec: Rigorous DTO Validation in API Gateway

## Problem Statement
The API Gateway currently performs minimal manual validation of incoming data (DTOs). This can lead to invalid data reaching downstream services or causing unexpected errors. We need a robust, consistent, and descriptive validation mechanism to reject malformed requests at the entry point.

## Proposed Solution
We will use the `github.com/go-playground/validator/v10` library to implement struct-tag-based validation for all DTOs used in `POST`, `PUT`, and `PATCH` endpoints in the API Gateway.

### Architecture

1.  **Validation Utility (`pkg/validator`):**
    *   Initialize a single `validator.Validate` instance.
    *   Provide a helper function to translate validation errors into user-friendly strings (e.g., "name is required" instead of "Key: 'CreateSensorRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag").

2.  **DTO Tags (`internal/types`):**
    *   Add `validate` tags to:
        *   `CreateSensorRequest`: `name` (required, min=3, max=100), `location` (max=255), `description` (max=500), `sensor_type_id` (required, gt=0).
        *   `UpdateSensorRequest`: optional fields but with length/value constraints if provided.
        *   `StoreReadingRequest`: `sensor_id` (required, gt=0), `value` (required), `timestamp` (optional).

3.  **Handler Integration (`services/api-gateway/handlers`):**
    *   Update `SensorHandler.CreateSensor` and `SensorHandler.UpdateSensor`.
    *   Update `WebSocketHandler.StoreReading`.
    *   The validation logic will:
        1. Decode JSON into the struct.
        2. Call `validator.Struct(req)`.
        3. If errors occur, return `400 Bad Request` with a JSON body containing descriptive error messages.

### Data Flow
1.  Client sends `POST /api/sensors`.
2.  `API Gateway` receives request.
3.  `SensorHandler.CreateSensor` decodes JSON.
4.  `Validator` checks tags.
5.  If invalid: `400 Bad Request` + `{"error": "name is required"}`.
6.  If valid: request proceeds to gRPC call.

### Error Handling
Errors will be returned as a JSON object:
```json
{
  "error": "Validation failed",
  "details": {
    "name": "name must be at least 3 characters",
    "sensor_type_id": "sensor_type_id must be greater than 0"
  }
}
```

### Testing Plan
*   **Unit Tests:** Add tests in `pkg/validator` for translation logic.
*   **Integration Tests:**
    *   Mock gRPC clients in handler tests.
    *   Send requests with missing/invalid fields and assert `400 Bad Request` and specific error messages.
    *   Send valid requests and assert `201 Created` / `200 OK`.
