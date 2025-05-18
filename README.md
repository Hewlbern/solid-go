# SolidGo - A Solid Protocol Server in Go

SolidGo is a high-performance implementation of the [Solid Protocol](https://solidproject.org/TR/protocol) using Go 1.24 and its standard library. This server provides a way to store and manage linked data with robust authentication and access control.

## Architecture

The server follows these design patterns:
- **Facade Pattern**: Simplifies the server API through a unified interface
- **Strategy Pattern**: For pluggable authentication methods, storage backends, and query handling
- **Factory Method**: Creates different types of Solid resources
- **Observer Pattern**: Handles events in the Solid server

### High-Level Components

```mermaid
graph TB
    Client[HTTP Client] --> Facade[Server Facade]
    Facade --> Auth[Authentication]
    Facade --> LDP[LDP Container Handler]
    Facade --> WAC[Web Access Control]
    
    Auth --> Auth1[WebID Strategy]
    Auth --> Auth2[OIDC Strategy]
    
    LDP --> Storage[Storage Interface]
    WAC --> Storage
    
    Storage --> Memory[Memory Storage]
    Storage --> FileSystem[File System Storage]
    
    Events[Event System] --> Observers[Event Observers]
    Facade --> Events
```

### Directory Structure

```mermaid
graph TB
    Root["/"] --> Cmd["cmd/"]
    Root --> Internal["internal/"]
    Root --> Pkg["pkg/"]
    Root --> Test["test/"]
    
    Cmd --> Server["server/"]
    
    Internal --> Auth["auth/"]
    Internal --> LDP["ldp/"]
    Internal --> RDF["rdf/"]
    Internal --> Storage["storage/"]
    Internal --> WAC["wac/"]
    Internal --> Events["events/"]
    Internal --> ServerImpl["server/"]
    
    Pkg --> Solid["solid/"]
    
    Test --> Smoke["smoke/"]
```

## Key Features

- Complete Solid Protocol implementation
- Standard library only - no external dependencies
- WebID and OIDC authentication support
- Web Access Control (WAC) for fine-grained access control
- Linked Data Platform (LDP) container management
- RDF parsing and serialization
- Event system for real-time updates
- Modular design with clean interfaces
- Containerized deployment with Docker

## Running Locally

```bash
# Build and run locally
go run cmd/server/main.go

# Run tests
go test ./...

# Run with Docker
docker-compose up -d
```

## API Endpoints

The server implements the standard Solid Protocol endpoints:

- `GET /`: Server information
- `GET /{container}/`: List resources in a container
- `POST /{container}/`: Create a new resource
- `GET /{resource}`: Get a resource
- `PUT /{resource}`: Create or update a resource
- `PATCH /{resource}`: Update a resource
- `DELETE /{resource}`: Delete a resource
- `HEAD /{resource}`: Get resource metadata
- `OPTIONS /{resource}`: Get resource options

## Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Auth as Authentication
    participant Storage
    
    Client->>Server: Request resource
    Server->>Auth: Authenticate request
    Auth->>Auth: Select auth strategy
    Auth->>Storage: Verify credentials
    Storage->>Auth: Authentication result
    Auth->>Server: Auth result
    
    alt is authenticated
        Server->>Client: Resource response
    else is not authenticated
        Server->>Client: 401 Unauthorized
    end
```

## Resource Management Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant LDP
    participant ACL as Access Control
    participant Storage
    
    Client->>Server: Request action on resource
    Server->>ACL: Check permissions
    ACL->>Storage: Fetch ACL rules
    Storage->>ACL: ACL data
    
    alt has permission
        ACL->>Server: Permission granted
        Server->>LDP: Handle resource operation
        LDP->>Storage: Perform storage operation
        Storage->>LDP: Operation result
        LDP->>Server: Operation result
        Server->>Client: Response
    else no permission
        ACL->>Server: Permission denied
        Server->>Client: 403 Forbidden
    end
```

## License

APACHE-2