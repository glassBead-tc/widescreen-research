# Ulysses Protocol

High-stakes debugging and problem-solving framework that prevents endless iteration cycles while maintaining quality through systematic phases and decision gates.

## Variables

PROBLEM_STATEMENT: $ARGUMENTS
ITERATION_LIMIT: $ARGUMENTS (default: 3)
CONFIDENCE_THRESHOLD: $ARGUMENTS (default: 0.95)
STAKES: "low" | "medium" | "high" | "critical"
COMMENTS: $ARGUMENTS

## Protocol Phases

### Phase 1: Reconnaissance (Time-boxed: 25% of budget)

```
OBJECTIVE: Understand the problem space completely

=> [swarm_intelligence](../orchestration/swarm-intelligence.md): Deploy research_agent to gather all available context
=> [swarm_intelligence](../orchestration/swarm-intelligence.md): Deploy analysis_agent to map system state and dependencies
=> [time_machine](../analysis/time-machine.md): Historical context and decisions
=> [knowledge_graph](../analysis/knowledge-graph.md): Visualize relationships

GATE: Can we clearly define the problem?
- [ ] Problem statement is specific and measurable
- [ ] Root cause hypotheses are formed
- [ ] Success criteria are defined
- [ ] Risk assessment is complete

If GATE fails: Escalate or change approach
```

### Phase 2: Strategic Planning (Time-boxed: 15% of budget)

```
OBJECTIVE: Design solution approach with multiple contingencies

-> [implementation_variants](../development/implementation-variants.md): Generate 3 solution approaches
-> [swarm_intelligence](../orchestration/swarm-intelligence.md): Evaluate approaches across dimensions
-> [pattern_synthesizer](../synthesis/pattern-synthesizer.md): Learn from similar problems

DECISION MATRIX:
- Implementation complexity
- Risk of unintended consequences
- Resource requirements
- Probability of success
- Rollback difficulty

GATE: Do we have a viable plan?
- [ ] Primary approach selected with high confidence
- [ ] Backup approaches identified
- [ ] Risk mitigation strategies in place
- [ ] Success metrics defined

If GATE fails: Return to reconnaissance or escalate
```

### Phase 3: Controlled Implementation (Time-boxed: 45% of budget)

```
OBJECTIVE: Execute solution with continuous validation

For each iteration (max 3):
  -> Implement smallest testable change
  -> Validate against success criteria
  -> Assess unintended consequences
  -> Document findings

  ITERATION_GATE:
  - [ ] Progress toward objective
  - [ ] No regression introduced
  - [ ] Within quality thresholds
  - [ ] Learning captured

  If ITERATION_GATE fails:
    - Try backup approach
    - Reduce scope
    - Escalate if at iteration limit

PHASE_GATE: Is solution working?
- [ ] Primary objectives met
- [ ] No critical regressions
- [ ] Quality maintained
- [ ] Monitoring in place

If PHASE_GATE fails: Activate contingency plan
```

### Phase 4: Validation & Documentation (Time-boxed: 15% of budget)

```
OBJECTIVE: Ensure solution is robust and knowledge is captured

-> [systematic_debug](../debugging/systematic-debug.md): Comprehensive testing
-> [evolution_tracker](../analysis/evolution-tracker.md): Update historical context
-> [pattern_synthesizer](../synthesis/pattern-synthesizer.md): Extract reusable patterns
-> mem0: Store decision rationale and lessons

FINAL_GATE: Is solution production-ready?
- [ ] All tests passing
- [ ] Performance verified
- [ ] Documentation updated
- [ ] Monitoring configured
- [ ] Rollback plan tested

If FINAL_GATE fails: Return to implementation or accept partial solution
```

## Decision Framework

### When to Continue vs. Stop

**Continue if:**

- Clear progress toward objectives
- No critical system damage
- Within iteration/time budget
- Learning is occurring

**Stop and Escalate if:**

- No progress after 2 iterations
- Critical system damage risk
- Problem scope expanding uncontrollably
- Stakes exceed available resources

**Accept Partial Solution if:**

- Core objective achieved (even if incomplete)
- Further iteration has diminishing returns
- Business/time constraints require it
- Sufficient foundation for future work

### Escalation Triggers

**Technical Escalation:**

- Problem requires expertise outside team
- Infrastructure/architectural changes needed
- Cross-team coordination required

**Management Escalation:**

- Resource constraints preventing solution
- Business priority conflicts
- Risk exceeds acceptable thresholds

## Quality Gates

### Code Quality

- No new linting errors
- Test coverage maintained
- Performance within bounds
- Security review passed

### System Quality

- No service degradation
- Error rates within limits
- Monitoring alerts clear
- Dependency health good

### Process Quality

- Documentation updated
- Knowledge captured in mem0
- Rollback procedures tested
- Team informed of changes

## Example Usage

```bash
/ulysses-protocol "Fix MCP telemetry integration causing agent workflow completion issues"

Stakes: HIGH (affects core product functionality)
Budget: 2 days
Iteration Limit: 3

Phase 1 (Reconnaissance - 4 hours):
- Historical analysis of MCP integration
- Current system state mapping
- Problem reproduction verification
- Risk assessment

Phase 2 (Planning - 2 hours):
- Multiple fix approaches identified
- Risk vs benefit analysis
- Rollback strategy defined

Phase 3 (Implementation - 12 hours):
- Iteration 1: Minimal fix attempt
- Iteration 2: Comprehensive approach
- Iteration 3: Fallback solution

Phase 4 (Validation - 2 hours):
- Full regression testing
- Performance validation
- Documentation updates
```

## Anti-Patterns to Avoid

### The Endless Debugging Spiral

- No clear success criteria
- No time limits
- No learning capture
- No escalation triggers

### The Silver Bullet Fallacy

- Assuming one approach will work
- No backup plans
- Over-engineering solutions
- Ignoring constraints

### The Hero Pattern

- One person solving everything
- No knowledge sharing
- No systematic approach
- No process improvement

## Meta-Learning

The Ulysses Protocol learns from each application:

- Refine gate criteria based on outcomes
- Improve estimation accuracy
- Better risk assessment
- Enhanced escalation triggers

This creates a feedback loop where the protocol becomes more effective at preventing unproductive work while maintaining solution quality.

## Integration with Other Commands

```bash
# Research phase
/time-machine "problematic-component"  # See: ../analysis/time-machine.md
/knowledge-graph "system-architecture"  # See: ../analysis/knowledge-graph.md

# Planning phase
/implementation-variants "potential solutions"  # See: ../development/implementation-variants.md
/swarm-intelligence "evaluate solution approaches"  # See: ../orchestration/swarm-intelligence.md

# Implementation phase
/parallel-explorer "test multiple approaches"  # See: ../exploration/parallel-explorer.md
/systematic-debug "validate each iteration"  # See: ../debugging/systematic-debug.md

# Validation phase
/pattern-synthesizer "extract learnings"  # See: ../synthesis/pattern-synthesizer.md
/context-aware-review "final quality check"  # See: ../analysis/context-aware-review.md
```

The Ulysses Protocol transforms debugging from reactive struggle into systematic problem-solving that builds organizational capability over time.

## Quick Reference: Workflow Commands

### Analysis & Research

- **[time_machine](../analysis/time-machine.md)** - Navigate through code history with context from mem0 entries
- **[knowledge_graph](../analysis/knowledge-graph.md)** - Advanced knowledge graph construction using multi-agent orchestration
- **[evolution_tracker](../analysis/evolution-tracker.md)** - Track and analyze system evolution over time
- **[context-aware-review](../analysis/context-aware-review.md)** - Comprehensive code and system review with context

### Development & Implementation

- **[implementation_variants](../development/implementation-variants.md)** - Generate multiple implementation approaches using different patterns
- **[parallel-executor](../development/parallel-executor.md)** - Execute multiple development tasks in parallel

### Debugging & Testing

- **[systematic_debug](../debugging/systematic-debug.md)** - Multi-phase debugging approach with instrumentation and hypothesis testing

### Orchestration & Coordination

- **[swarm_intelligence](../orchestration/swarm-intelligence.md)** - Deploy specialized AI agents across multiple contexts
- **[meta-orchestrator](../orchestration/meta-orchestrator.md)** - High-level orchestration of complex workflows

### Exploration & Discovery

- **[parallel-explorer](../exploration/parallel-explorer.md)** - Explore multiple solution paths simultaneously

### Synthesis & Learning

- **[pattern_synthesizer](../synthesis/pattern-synthesizer.md)** - Extract, analyze, and synthesize patterns across codebases

### Research Capabilities

- **[intelligent-mcp-research-suite](../research/intelligent-mcp-research-suite.md)** - Comprehensive research capabilities using MCP tools
- **[advanced-retrieval-orchestrator](../research/advanced-retrieval-orchestrator.md)** - Advanced information retrieval and analysis

Each workflow is designed to integrate seamlessly with the Ulysses Protocol phases, providing systematic approaches to complex problem-solving while maintaining quality and preventing endless iteration cycles.
