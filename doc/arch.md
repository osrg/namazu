# Earthquake Archtecture

# Overview

    +----------+     +---------------+     +--------------------------------------------------+
    |  Testee  | --- | EQ Inspector  | --- |                                                  |
    +----------+     +---------------+     |                  EQ Orchestrator                 |				 
                                           |                                                  |
    +----------+     +---------------+     |   * Receives events from the inspectors	      |
    |  Testee  | --- | EQ Inspector  | --- |   * Generate action set from events              |
    +----------+     +---------------+     |     (Including injected fault actions)           |
                                           |   * Send actions to the inspectors               |
    +----------+     +---------------+     |     in various orders                            |
    |  Testee  | --- | EQ Inspector  | --- |                                                  |
    +----------+     +---------------+     +--------------------------------------------------+
                  ^                     ^
                  |                     |
                  |                     +--- JSON over HTTP (Inspectors POSTs events, then GETs actions)
                  |
                  +--- Inspectors (for Java) are embedded in testee programs using byteman


## Inspectors
### Implemented Inspectors:
 * Java: byteman
 * Ethernet: Open vSwitch + ryu
 
## Orchestrator
### Implemented Backend and Bindings:
 * `libearthquake`: backend library for recording execution history.
 * `pyearthquake`: python binding for `libearthquake`

### Implemented Plug-ins:
 * Orchestrator Plug-in: manages explorer and so on.
  * `BasicOrchestrator`: a basic orchestrator
  
 * Explorer Plug-in: explores state space
  * `DumbExplorer`: reorder nothing
  * `RandomExplorer`: reorder actions in randomly
  * `GreedyExplorer` (WIP)
  
 * Process Watcher Plug-in: watches events from processes, maps event->action, and execute actions. You can check safety properties(i.e. assertion) in Process Watcher.
  * 'BasicProcessWatcher`: a basic process watcher
  
 * Termination Detector Plug-in:  detects termination of one-shot execution. Execution histories are recorded to storage on such terminations.
  * `InspectionEndDetector`: detects termination when `InspectionEndEvent`s are observed from all processes
  * `IdleForWhileDetector`: detects termination when no event are observed for several milliseconds.

### Implemented API
 * Orchestartor provides RESTful (JSON over HTTP) API.
  * HTTP enables simplification of connection handling and proxying
   * Even if inspectors crashed unexpectedly, you can send `InspectionEndEvent`s to the orchestrator manually with `curl`/`wget`.
  * JSON enables flexible structuralization. Some JSON-friendly tools (e.g., JQ, MongoDB, ..) can be used with execution history JSONs.
 * `POST /api/v1` (Non-blocking): send an event to the orchestrator
 * `GET /api/v1/<process_id>` (Blocking): receive an action for <process_id>from the orchestrator.
 * `GET /visualize_api/csv`: GET CSV statistics. You may use `gnuplot` to visualize CSV.

### Implemented Entities
 * Events:
  * `FunctionCallEvent`: inspected and deferred function calls
  * `PacketEvent`: inspected and deferred ethernet packets
  * `LogEvent`: inspected syslog (WIP)
  * `InspectionEndEvent`: termination of inspectors. usable for the `InspectionEndDetector` plug-in.
  
 * Actions:
  * `NopAction`: nop
  * `PassDeferredEventAction`: for `FunctionCallEvent` and `PacketEvent`
  * `ExecuteCommandOnInspectorAction`: (WIP)
  * `ExecuteCommandOnOrchestratorAction`: usable for fault-injection that kills inspectors



 