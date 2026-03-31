# Golaude

**Golaude** is a high-performance, CLI-based AI coding assistant built in Go. It implements a fully autonomous **Agent Loop** that allows an LLM to interact directly with your local file system and terminal through OpenAI-compatible tool calling. This project was built following the [**CodeCrafters "Build Your Own Claude Code" Challenge**](https://app.codecrafters.io/courses/claude-code/overview) and served as a deep dive into tool-assisted LLM reasoning, Go's os/exec and filesystem primitives, and stateless API design.

[![progress-banner](https://backend.codecrafters.io/progress/claude-code/efd74af4-bc4b-479f-b9cb-b51c4a9ba843)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)


## Features

* **Recursive Agent Loop**: Implements a sophisticated "Reason-Act-Observe" cycle, allowing the assistant to chain multiple tool calls together to solve complex tasks.
* **System Integration**: Built-in tools for safe file system operations (Read, Write) and terminal command execution via bash.
* **Stateless History Management**: Manages a continuous conversation transcript, ensuring the LLM maintains context across multiple rounds of tool execution.
* **Go-Native Performance**: Leverages Go’s strong typing and concurrency primitives for a fast, reliable developer experience.
