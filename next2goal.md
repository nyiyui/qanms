# Goal-based configuration

## Goals
- reduce state in unncessary places (e.g. package central)

## Design
When CS wants to change the network:
1. CS sends a `goal.Machine` to each machine it wants to update
2. Nodes implement the goal (using `goal.ApplyMachineDiff`) and report success/failure
  - Note: Nodes could report completion per-network
3. CS waits for all goals to be implemented
