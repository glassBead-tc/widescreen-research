# Target Bidirectional Architecture Specification

**Date:** October 6, 2025
**Status:** Design Complete

---

## Executive Summary

This document outlines the target architecture for the research system, introducing a new host application, `widescreen-research`, which serves as a bidirectional Model Context Protocol (MCP) entity. This entity acts as an MCP server to external clients (like Claude Desktop) and an MCP client to the dedicated GCP orchestration server (`widescreen-research-mcp`). This design centralizes research session management and report aggregation within the host application, while delegating heavy orchestration tasks (Cloud Run deployment, Pub/Sub management) to the server-only `widescreen-research-mcp`.

---

## Architectural Components

The system is composed of three main layers:

### 1. Client Layer: Claude Desktop

- **Role:** Initiates research workflows and consumes final reports.
- **Interface:** MCP Client (stdio/network) connecting to `widescreen-research`.
- **Responsibility:** Calls the `orchestrate-research` tool exposed by `widescreen-research`. Receives a notification when the master report is ready, and calls a resource tool to retrieve the report content.

### 2. Host Layer: `widescreen-research` (Bidirectional MCP Entity)

- **Role:** Central coordination, session state management, report aggregation, and bidirectional MCP gateway.
- **Interfaces:**
  - **MCP Server:** Exposes research tools (e.g., `orchestrate-research`, `get-report`) to Claude Desktop.
  - **MCP Client:** Connects to `widescreen-research-mcp` to delegate GCP orchestration tasks.
- **Key Responsibilities:**
  - Receives `orchestrate-research` request from Claude Desktop.
  - Manages the overall research session lifecycle.
  - Calls `widescreen-research-mcp`'s orchestration tool.
  - Receives and processes collected drone data objects from `widescreen-research-mcp` (via a queue/callback mechanism).
  - Aggregates drone data into a master report.
  - Sends completion notification to Claude Desktop.

### 3. Orchestration Layer: `widescreen-research-mcp` (Server-Only)

- **Role:** Dedicated GCP resource provisioning and management.
- **Interface:** MCP Server (stdio/network) connected to `widescreen-research`.
- **Key Responsibilities (Unchanged from Post-Slop Cleanup):**
  - Receives orchestration commands from `widescreen-research`.
  - Spins up N Cloud Run instances (Drones).
  - Manages Pub/Sub topics for result collection.
  - Sends tasks to Drones (HTTP POST).
  - Collects raw results from Drones via Pub/Sub and forwards them back to `widescreen-research` for aggregation.

### 4. Drone Layer: Cloud Run Agents (N Instances)

- **Role:** Execute individual, identical agentic research workflows on specific entities.
- **Communication:** Receives tasks via HTTP, publishes results to Pub/Sub.

---

## Data Flow and Component Relationships

### 1. Research Initiation (Claude Desktop → `widescreen-research`)

1. Claude Desktop (MCP Client) calls `widescreen-research` (MCP Server) with the `orchestrate-research` tool.
2. `widescreen-research` creates a new research session state.

### 2. Orchestration Delegation (`widescreen-research` → `widescreen-research-mcp`)

1. `widescreen-research` (MCP Client) calls `widescreen-research-mcp` (MCP Server) with a command to provision resources and start the drone research process.
2. `widescreen-research-mcp` deploys Cloud Run Drones and sets up Pub/Sub topics.

### 3. Asynchronous Result Collection (Drones → `widescreen-research-mcp` → `widescreen-research`)

1. Drones execute research and publish `DroneResult` objects to Pub/Sub.
2. `widescreen-research-mcp`'s orchestrator collects these results via its Pub/Sub queue.
3. **CRITICAL CHANGE:** Instead of generating the final report, `widescreen-research-mcp` streams or sends the collected `DroneResult` objects back to `widescreen-research` (Host Layer) concurrently. This transfer mechanism (e.g., a dedicated MCP tool call back to the host, or a separate Pub/Sub topic monitored by the host) must be defined.

### 4. Report Aggregation and Notification (`widescreen-research` → Claude Desktop)

1. `widescreen-research` receives all `DroneResult` objects.
2. `widescreen-research` processes the results concurrently, aggregates them into a `ResearchReport`, and stores it.
3. `widescreen-research` sends a notification back to Claude Desktop (e.g., via an MCP response or a separate notification channel) indicating the report is ready.

### 5. Report Retrieval (Claude Desktop → `widescreen-research`)

1. Claude Desktop calls the `get-report` tool (exposed by `widescreen-research`) or accesses the report as an embedded resource.

---

## Integration Points and Interfaces

| Component | Interface | Purpose |
| :--- | :--- | :--- |
| **Claude Desktop** | MCP Client (stdio) | Calls `orchestrate-research` on Host Layer. |
| **`widescreen-research`** | MCP Server (stdio) | Exposes `orchestrate-research` and `get-report` tools. |
| **`widescreen-research`** | MCP Client (stdio/network) | Calls `start-gcp-orchestration` on Orchestration Layer. |
| **`widescreen-research-mcp`** | MCP Server (stdio/network) | Exposes `start-gcp-orchestration` tool. |
| **`widescreen-research-mcp`** | GCP SDKs | Manages Cloud Run, Pub/Sub, Firestore. |
| **Drones** | HTTP POST | Receives initial task instructions from Orchestration Layer. |
| **Drones** | GCP Pub/Sub | Publishes raw `DroneResult` objects to Orchestration Layer. |

---

## Required Changes to `widescreen-research-mcp`

The `widescreen-research-mcp` repository remains a server-only component, but its `OrchestrateResearch()` method must be modified:

1. **Rename/Refactor:** The primary tool should be renamed from `orchestrate-research` (which implies full end-to-end control) to something like `start-gcp-orchestration` or `deploy-and-monitor-drones`.
2. **Remove Report Generation:** The `generateReport()` and `analyzeResults()` steps (lines 215-216 in the previous workflow) must be removed from `widescreen-research-mcp`.
3. **Result Forwarding:** The collected results must be forwarded back to the calling `widescreen-research` host application. This requires `widescreen-research-mcp` to accept a callback endpoint or use a dedicated Pub/Sub topic for results that the host application monitors.

---

## Architectural Diagram (Mermaid)

```mermaid
graph TD
    subgraph Client Layer
        CD[Claude Desktop]
    end

    subgraph Host Layer (widescreen-research)
        direction LR
        H_S[MCP Server]
        H_C[MCP Client]
        H_A(Report Aggregator & State Manager)
        H_N(Notification Sender)
        H_S -- 1. orchestrate-research --> H_A
        H_A -- 2. start-gcp-orchestration --> H_C
        H_A -- 4. Master Report Ready --> H_N
        H_N -- 5. Notification --> CD
        CD -- 6. get-report --> H_S
    end

    subgraph Orchestration Layer (widescreen-research-mcp)
        direction LR
        O_S[MCP Server]
        O_D(Drone Deployer)
        O_Q(Pub/Sub Result Collector)
        O_F(Result Forwarder)
        O_S -- 2. start-gcp-orchestration --> O_D
        O_D -- 3. Deploy Drones --> DRONES
        DRONES -- 3. Pub/Sub Results --> O_Q
        O_Q -- 3. Forward Results --> O_F
        O_F -- 3. Stream Results --> H_A
    end

    subgraph Drone Layer (Cloud Run Agents)
        DRONES[N Cloud Run Instances]
    end

    H_C -- stdio/network --> O_S
    O_F -- network/callback --> H_A
