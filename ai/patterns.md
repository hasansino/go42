## Subagent Architecture Patterns

### 1. Specialized Domain Experts

• Create agents with deep expertise in specific areas (legal, medical, technical, creative)
• Each agent has its own specialized prompt with domain-specific instructions
• Use a router/orchestrator to delegate tasks to appropriate experts
• Consider overlapping expertise for cross-domain problems

### 2. Task-Based Decomposition

• Research Agent: Focuses on gathering and synthesizing information
• Analysis Agent: Performs deep analytical thinking and reasoning
• Creative Agent: Handles brainstorming, ideation, and creative solutions
• Validator Agent: Reviews and fact-checks outputs from other agents
• Editor Agent: Refines and polishes final outputs

### 3. Cognitive Function Separation

• Perception Agent: Processes and interprets input data
• Memory Agent: Manages context and retrieves relevant information
• Reasoning Agent: Handles logical deduction and problem-solving
• Decision Agent: Makes choices based on analyzed options
• Communication Agent: Formats and delivers responses

## Implementation Strategies

### 4. Hierarchical Structure

• Master Coordinator: High-level planning and delegation
• Middle Managers: Domain or task-specific coordinators
• Worker Agents: Specialized execution units
• Consider feedback loops between levels

### 5. Peer-to-Peer Network

• Agents communicate directly without central authority
• Consensus mechanisms for decision-making
• Emergent behavior from agent interactions
• Self-organizing teams based on task requirements
