# unravel-be
Managing events and entities


# Project Structure
This project follows the clean architecture.

```
unravel-be/
├── cmd/                # Frameworks & Drivers
│   └── main.go         # Application entry point
├── config/             # Frameworks & Drivers
│   └── config.go
│   └── ...
├── application/                # Contains core application logic and reusable components
│   ├── domain/         # Entities (Core Business Rules - independent of any framework)
│   │   ├── parameter.go
│   │   ├── schema.go
│   │   └── ...
│   │
│   ├── usecase/        # Application-specific Business Rules (Application Layer)
│   │   ├── parameter.go   # Defines ParameterService interface, ParameterRepository interface
│   │   ├── schema.go      # Defines SchemaService interface, SchemaRepository interface
│   │   └── ...
│   │
│   └── adapter/        # Interface Adapters Layer
│       ├── persistence/ # Database Repositories (implements usecase's repository interfaces)
│       │   └── postgres/
│       │       ├── parameter_repository.go
│       │       ├── schema_repository.go
│       │       └── ...
│       │
│       ├── delivery/    # Web/API Handlers (calls usecase services)
│       │   ├── http/
│       │   │   ├── server.go
│       │   │   ├── router.go
│       │   │   ├── parameter/
│       │   │   │   ├── handler.go
│       │   │   │   ├── model.go
│       │   │   │   └── route.go
│       │   │   └── schema/
│       │   │       ├── handler.go
│       │   │       ├── model.go
│       │   │       └── route.go
│       │   └── grpc/  # If you have gRPC
│       │       └── ...
│       │
│       └── external/    # Clients for external APIs/services
│           └── ...
│
├── db/                 # Database related (migrations, schema tools)
│   ├── migration/
│   │   ├── 000001_create_parameters_table.up.sql
│   │   ├── 000001_create_parameters_table.down.sql
│   │   └── ...
│   ├── .gen/           # Jet generated code
│   └── ...
└── go.mod
└── go.sum
└── makefile
└── Dockerfile
└── docker-compose.yml
└── deployment.yaml
└── README.md
```

# API Reference

## Health Check

### GET /health
Returns the health status of the service.

**Response:**
```json
{
  "message": "ok"
}
```

## Parameters

### GET /parameters
Returns all parameters stored in the database.

**Response:**
```json
{
  "data": [
    {
      "parameter_key": "string",
      "data_type": "string"
    }
  ]
}
```

## Schemas

### POST /schemas
Create a new schema. A unique version is automatically generated based on the parameters provided.

**Request Body:**
```json
{
  "name": "string",
  "parameters": {
    "key": "value"
  }
}
```

**Response (201 Created):**
```json
{
  "id": "uuid",
  "name": "string",
  "version": "v1-<hash>",
  "parameters": {},
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or parameters
- `409 Conflict`: Schema version already exists

### GET /schemas/:name
Returns all versions of a schema by name.

**Path Parameters:**
- `name`: The name of the schema

**Response (200 OK):**
```json
[
  {
    "id": "uuid",
    "name": "string",
    "version": "v1-<hash>",
    "parameters": {},
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
]
```

**Error Responses:**
- `404 Not Found`: Schema not found

### GET /schemas/:name/versions/:version
Returns a specific version of a schema.

**Path Parameters:**
- `name`: The name of the schema
- `version`: The version identifier (e.g., `v1-abc12345`)

**Response (200 OK):**
```json
{
  "id": "uuid",
  "name": "string",
  "version": "v1-<hash>",
  "parameters": {},
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Error Responses:**
- `404 Not Found`: Schema or version not found
