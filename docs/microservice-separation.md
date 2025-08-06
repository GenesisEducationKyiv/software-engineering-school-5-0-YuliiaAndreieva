# Microservice Architecture Design

---

## Implementation Consideration

During the implementation phase, I will test and evaluate whether **gRPC** or **HTTP/REST** provides better performance for service-to-service communication for my case.
The current architecture assumes gRPC for internal communication, but I maintain flexibility to switch to HTTP/REST if testing shows it's more suitable for our scale and requirements.

---

## Overview

This document describes the proposed microservice architecture for the weather subscription system, breaking down the monolithic application into focused, independently deployable services.

---

## Microservice Description

### 1. API Gateway
**Purpose**: Single entry point for all client requests

**Responsibilities**:
- Route HTTP requests to appropriate microservices
- Handle request/response transformation

**Endpoints**:
- `GET /weather` - Weather information
- `POST /subscribe` - Create subscription
- `GET /confirm/{token}` - Confirm subscription
- `GET /unsubscribe/{token}` - Remove subscription

---

### 2. Subscription Service
**Purpose**: Core business logic for subscription management

**Responsibilities**:
- Create, read, update, delete subscriptions
- Provide subscription data for broadcasting
- Serve static HTML page for subscription interface

**Core Functions**:
- Subscription CRUD operations
- Data persistence

**Communication**:
- Publishes events to RabbitMQ for email notifications
- Receives gRPC calls from API Gateway
- Provides gRPC endpoints for Weather Broadcast Service
- gRPC calls to Token Service for token operations

**Technology Stack**:
- PostgreSQL for data persistence
- gRPC server/client
- RabbitMQ publisher

---

### 3. Token Service
**Purpose**: Token generation, validation

**Responsibilities**:
- Generate secure JWT tokens for subscription confirmation
- Validate token authenticity and expiration

**Core Functions**:
- JWT token generation with configurable expiration
- Token validation and verification (signature, expiration, structure)

**Token Storage Strategy**:
- **JWT tokens are stateless** - no permanent storage required
- **Subscription Service** stores tokens in PostgreSQL with subscription data

**Communication**:
- Receives gRPC calls from Subscription Service for token generation
- Receives gRPC calls from API Gateway for token validation
- Provides gRPC endpoints for token operations

**Technology Stack**:
- Go with JWT library
- gRPC server

---

### 4. Weather Service
**Purpose**: External weather data integration and caching

**Responsibilities**:
- Fetch weather data from external APIs
- Cache weather results to reduce API calls
- Provide weather data to other services

**Core Functions**:
- Weather data retrieval from WeatherAPI.com and OpenWeatherMap
- Redis caching for weather results
- City validation against external API

**Communication**:
- HTTP calls to external weather APIs
- gRPC server for internal service communication
- Redis for caching

**Technology Stack**:
- Go with weather API integration
- Redis for caching
- gRPC server
- HTTP client for external APIs

---

### 5. Weather Broadcast Service
**Purpose**: Scheduled weather updates delivery

**Responsibilities**:
- Execute scheduled weather broadcasts (hourly/daily)
- Coordinate between services for data collection
- Manage broadcast scheduling and execution
- Handle broadcast failures and retries

**Core Functions**:
- Scheduled job execution (cron-based)
- Subscription data collection
- Weather data aggregation
- Email notification triggering

**Communication**:
- gRPC calls to Subscription Service for active subscriptions
- gRPC calls to Weather Service for weather data
- Publishes events to RabbitMQ for email delivery
- Receives scheduling triggers

**Technology Stack**:
- Go with cron scheduling
- gRPC client for service communication
- RabbitMQ publisher
- Job queue management

---

### 6. Email Service
**Purpose**: Email delivery and management

**Responsibilities**:
- Send confirmation emails
- Send weather update emails
- Handle email delivery failures
- Manage email templates
- Validate email addresses during sending

**Core Functions**:
- SMTP integration
- Email template rendering
- Delivery status tracking
- Email validation and bounce handling
- Email existence verification during delivery

**Communication**:
- Consumes events from RabbitMQ
- SMTP communication for email delivery
- Event publishing for delivery status

**Technology Stack**:
- Go with SMTP integration
- RabbitMQ consumer
- Email template engine
- SMTP client

---

## Communication Patterns

### gRPC
- API Gateway ↔ Subscription Service
- API Gateway ↔ Weather Service
- API Gateway ↔ Token Service
- Subscription Service ↔ Token Service
- Weather Broadcast Service ↔ Subscription Service
- Weather Broadcast Service ↔ Weather Service

### RabbitMQ
- Subscription Service → Email Service (confirmation emails)
- Weather Broadcast Service → Email Service (weather updates)

### External Communication
- Weather Service → WeatherAPI.com and OpenWeatherMap (HTTP)
- Email Service → SMTP Server

---

## Service Communication Diagram

### High-Level Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        Client[Client/Browser]
    end

    subgraph "API Gateway"
        AG[API Gateway<br/>Gin + gRPC Client]
    end

    subgraph "Core Services"
        SS[Subscription Service<br/>]
        TS[Token Service<br/>]
        WS[Weather Service<br/>]
        WBS[Weather Broadcast Service<br/>]
        ES[Email Service<br/>]
    end

    subgraph "External Services"
        WeatherAPI[WeatherAPI.com]
        OpenWeather[OpenWeatherMap]
        SMTPServer[SMTP Server]
    end

    subgraph "Infrastructure"
        PG[(PostgreSQL)]
        Redis[(Redis)]
    end

    %% Client to API Gateway
    Client -->|HTTP| AG

    %% API Gateway to Services (gRPC)
    AG -->|gRPC| SS
    AG -->|gRPC| TS
    AG -->|gRPC| WS

    %% Service to Service (gRPC)
    SS -->|gRPC| TS
    WBS -->|gRPC| SS
    WBS -->|gRPC| WS

    %% Services to External APIs
    WS -->|HTTP| WeatherAPI
    WS -->|HTTP| OpenWeather
    ES -->|SMTP| SMTPServer

    %% Services to Infrastructure
    SS --> PG
    WS --> Redis

    %% Styling
    classDef clientLayer fill:#e1f5fe
    classDef gateway fill:#f3e5f5
    classDef coreService fill:#e8f5e8
    classDef externalService fill:#fff3e0
    classDef infrastructure fill:#fce4ec

    class Client clientLayer
    class AG gateway
    class SS,TS,WS,WBS,ES coreService
    class WeatherAPI,SMTPServer,OpenWeather externalService
    class PG,Redis infrastructure
```

### Detailed Communication Flows

#### 1. New Subscription Flow
```mermaid
sequenceDiagram
    participant Client
    participant AG as API Gateway
    participant SS as Subscription Service
    participant TS as Token Service
    participant PG as PostgreSQL
    participant RabbitMQ
    participant ES as Email Service
    participant SMTP

    Client->>AG: POST /subscribe
    AG->>SS: CreateSubscription(email, city, frequency)
    SS->>PG: Check if email already subscribed
    PG-->>SS: Subscription status
    alt Email not subscribed
        SS->>TS: GenerateToken(email, type: "confirmation")
        TS-->>SS: JWT Token
        SS->>PG: Create subscription with token
        SS->>RabbitMQ: Publish EmailEvent(confirmation)
        RabbitMQ->>ES: Consume EmailEvent
        ES->>SMTP: Send confirmation email
        alt Email delivery fails
            SMTP-->>ES: Email not delivered
            ES-->>RabbitMQ: Publish EmailFailureEvent
        else Email delivered successfully
            SMTP-->>ES: Email sent
            ES-->>RabbitMQ: Publish EmailSuccessEvent
        end
        SS-->>AG: Subscription created
        AG-->>Client: Success response
    else Email already subscribed
        SS-->>AG: Error: Email already subscribed
        AG-->>Client: 409 Conflict
    end
```

#### 2. Subscription Confirmation Flow
```mermaid
sequenceDiagram
    participant Client
    participant AG as API Gateway
    participant TS as Token Service
    participant SS as Subscription Service
    participant PG as PostgreSQL

    Client->>AG: GET /confirm/{token}
    AG->>TS: ValidateToken(token)
    TS-->>AG: Token validation result
    alt Token valid
        AG->>SS: ConfirmSubscription(token)
        SS->>PG: Update subscription status
        PG-->>SS: Updated
        SS-->>AG: Confirmed
        AG-->>Client: Success response
    else Token invalid
        AG-->>Client: 400 Bad Request
    end
```

#### 3. Weather Request Flow
```mermaid
sequenceDiagram
    participant Client
    participant AG as API Gateway
    participant WS as Weather Service
    participant Redis
    participant WeatherAPI as WeatherAPI.com
    participant OpenWeather as OpenWeatherMap

    Client->>AG: GET /weather?city={city}
    AG->>WS: GetWeather(city)
    WS->>Redis: Check cache
    alt Cache hit
        Redis-->>WS: Cached weather data
        WS-->>AG: Weather data
        AG-->>Client: JSON response
    else Cache miss
        WS->>WeatherAPI: Fetch weather data
        alt WeatherAPI returns city not found
            WeatherAPI-->>WS: City not found
            WS->>OpenWeather: Fetch weather data (fallback)
            alt OpenWeather returns city not found
                OpenWeather-->>WS: City not found
                WS-->>AG: Error: City not found
                AG-->>Client: 404 Not Found
            else OpenWeather returns weather
                OpenWeather-->>WS: Weather response
                WS->>Redis: Cache weather data
                WS-->>AG: Weather data
                AG-->>Client: JSON response
            end
        else WeatherAPI returns weather
            WeatherAPI-->>WS: Weather response
            WS->>Redis: Cache weather data
            WS-->>AG: Weather data
            AG-->>Client: JSON response
        end
    end
```

#### 4. Scheduled Weather Broadcast Flow
```mermaid
sequenceDiagram
    participant Cron
    participant WBS as Weather Broadcast Service
    participant SS as Subscription Service
    participant WS as Weather Service
    participant PG as PostgreSQL
    participant Redis
    participant RabbitMQ
    participant ES as Email Service
    participant SMTP
    participant WeatherAPI as WeatherAPI.com
    participant OpenWeather as OpenWeatherMap

    Cron->>WBS: Trigger scheduled broadcast
    WBS->>SS: GetActiveSubscriptions(frequency)
    SS->>PG: Query active subscriptions
    PG-->>SS: Subscription list
    SS-->>WBS: Active subscriptions
    
    loop For each subscription
        WBS->>WS: GetWeather(city)
        WS->>Redis: Check cache
        alt Cache hit
            Redis-->>WS: Cached weather
        else Cache miss
            WS->>WeatherAPI: Fetch weather
            alt WeatherAPI returns city not found
                WS->>OpenWeather: Fetch weather (fallback)
                alt OpenWeather returns city not found
                    OpenWeather-->>WS: City not found
                    WS-->>WBS: Error: City not found
                    Note over WBS: Skip this subscription
                else OpenWeather returns weather
                    OpenWeather-->>WS: Weather data
                    WS->>Redis: Cache weather
                end
            else WeatherAPI returns weather
                WeatherAPI-->>WS: Weather data
                WS->>Redis: Cache weather
            end
        end
        alt Weather data available
            WS-->>WBS: Weather data
            WBS->>RabbitMQ: Publish WeatherUpdateEvent
        else City not found
            Note over WBS: Skip weather update for this subscription
        end
    end
    
    RabbitMQ->>ES: Consume WeatherUpdateEvent
    ES->>SMTP: Send weather update email
```

### Communication Patterns Summary

**Service Abbreviations:**
- **AG** = API Gateway
- **SS** = Subscription Service  
- **TS** = Token Service
- **WS** = Weather Service
- **WBS** = Weather Broadcast Service
- **ES** = Email Service

| Service | Incoming gRPC | Outgoing gRPC | RabbitMQ Publisher | RabbitMQ Consumer | External APIs |
|---------|---------------|---------------|-------------------|-------------------|---------------|
| **Subscription Service** | AG, WBS | TS | Email events | - | - |
| **Token Service** | AG, SS | - | - | - | - |
| **Weather Service** | AG, WBS | - | - | - | WeatherAPI.com and OpenWeatherMap |
| **Weather Broadcast Service** | - | SS, WS | Weather events | - | - |
| **Email Service** | - | - | - | Email events | SMTP Server |

### Data Storage Patterns

| Service | Primary Storage | Cache | Message Queue |
|---------|----------------|-------|---------------|
| **Subscription Service** | PostgreSQL | - | RabbitMQ Publisher |
| **Token Service** | - | - | - |
| **Weather Service** | - | Redis | - |
| **Weather Broadcast Service** | - | - | RabbitMQ Publisher |
| **Email Service** | - | - | RabbitMQ Consumer |
