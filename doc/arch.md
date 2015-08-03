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
                  |                     +--- Protocol Buffers, or REST (Inspectors POSTs events, then GETs actions)
                  |
                  +--- Inspectors (for Java) are embedded in testee programs using byteman


Please also refer to [figs/eq-redesign.pdf](figs/eq-redesign.pdf) for planned archtecture-level changes.

## Inspectors

 * Java: byteman
 * Ethernet (ryu): Open vSwitch + ryu
 * Ethernet (nfqhook): iptables + NFQUEUE

 
## Orchestrator

Explore Policy: explores state space

  * `Dumb`: reorder nothing
  * `Random`: reorder actions randomly
  * `DFS`,`BFS`: (WIP)
  * `DPOR`: (WIP) enables dynamic partial order reduction

History Storage

 * `naive`
 * `mongodb`

### REST API

 * `POST /api/v2/events/<entity_id>/<event_uuid>` (Non-blocking): send an event to the orchestrator
 * `GET /api/v2/actions/<entity_id>` (Blocking): receive an action for <entity_id> from the orchestrator.
 * `DELETE /api/v2/actions/<entity_id>/<action_uuid>` (Non-blocking): ack for get

Events:

 * `FunctionCallEvent`: inspected and deferred function calls
 * `FunctionReturnEvent`: inspected and deferred function calls
 * `PacketEvent`: inspected and deferred Ethernet packets
 * `LogEvent`: inspected syslog

Actions:

 * `AcceptEventAction`: accept an event
 * `DropDeferredEventAction`: (planned)
 * `ExecuteCommandOnInspectorAction`: (planned)
 * `ExecuteCommandOnOrchestratorAction`: (planned)


### pyearthquake plug-ins (was available in v0.1, but not under active development)

 * Orchestrator Plug-in: manages Explorer and so on.
  * `BasicOrchestrator`: a basic orchestrator
  
 * Explorer Plug-in: explores state space
  * `DumbExplorer`: reorder nothing
  * `RandomExplorer`: reorder actions randomly
  * `TimeBoundedRandomExplorer`: similar to `RandomExplorer`, but maximum deferred time is bounded
  * `GreedyExplorer`: DFS-like policy using NetworkX graph processing library
  
 * Process (now called "Entity") Watcher Plug-in: watches events from processes, maps event->action, and execute actions. You can check safety properties (i.e. assertion) in Process Watcher.
  * `BasicProcessWatcher`: a basic process watcher
  
 * Termination Detector Plug-in:  detects termination of one-shot execution. Execution histories are recorded to storage on such terminations.
  * `InspectionEndDetector`: detects termination when `InspectionEndEvent`s are observed from all processes
  * `IdleForWhileDetector`: detects termination when no event are observed for several milliseconds.

