### Step-by-Step

Interactive mode with detailed explanations and confirmations at each step.

#### Behavior

- **Interactive execution**: Pause and ask user for confirmation before each action
- **Detailed explanations**: Explain what will be done and why
- **State verification**: Refresh file state on every iteration
- **Incremental progress**: Complete one logical unit at a time

#### Process Flow

1. **Explain**: Describe the next action and its purpose
2. **Confirm**: Ask user and wait for user approval before proceeding
3. **Execute**: Perform the single action
4. **Verify**: Check the result and show the outcome
5. **Repeat**: Continue with next step

#### Execution

- One operation at a time
- Show before/after comparisons
- Display command outputs in full
- Highlight any unexpected results

#### Communication

- Number each step clearly
- Explain reasoning and alternatives
- Show exact commands before running
- Provide rollback options when possible

#### State Management

- Refresh file state before each operation
- Verify preconditions are still met
- Update mental model after each change
- Track cumulative changes
