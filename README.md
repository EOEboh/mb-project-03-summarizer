**The rule:** AI communication only lives in `ai/`. Handlers never call `http.Post` to Ollama directly. This makes it trivially easy to swap providers.

---

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.com) installed and running (the Mac app auto-starts; run `ollama serve` only if using the CLI)

## Quick Start

```bash
# 1. Clone or use this as a GitHub template
git clone https://github.com/EOEboh/mb-bootcamp-scaffold my-project
cd my-project

# 2. Replace the module name in go.mod
# Change: github.com/EOEboh/mb-bootcamp-scaffold
# To:     github.com/EOEboh/my-project-name

# 3. Pull the required models (first time only — ~6 GB total)
make setup

# 4. Run
make run
# → http://localhost:8080
```

---

## The ai/ Package API

Every project uses exactly two functions:

```go
// Non-streaming: returns the full response as a string
response, err := ai.Chat(ai.DefaultModel, []ai.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user",   Content: "Hello!"},
})

// Streaming: calls onChunk for each token as it arrives
err := ai.ChatStream(ai.DefaultModel, messages, func(chunk string) error {
    fmt.Println(chunk) // do something with each token
    return nil         // return error to abort stream early
})
```

---

## Stack

| Layer      | Technology                |
|------------|---------------------------|
| Backend    | Go 1.22+                  |
| LLM        | Ollama (local)            |
| Frontend   | HTMX + Vanilla JS         |
| Database   | SQLite (projects 4, 5, 10)|
| Streaming  | SSE / WebSockets          |
| Deploy     | Docker (project 10)       |

---

## Projects Built on This Scaffold

| #  | Project                  | Key Addition                        | Repo                                                                                        |
|----|--------------------------|-------------------------------------|---------------------------------------------------------------------------------------------|
| 1  | AI Chat Interface        | SSE streaming                       | [mb-project-01-chat](https://github.com/EOEboh/mb-project-01-chat)                         |
| 2  | Code Snippet Explainer   | System prompts + code models        | [mb-project-02-code-explainer](https://github.com/EOEboh/mb-project-02-code-explainer)     |
| 3  | Smart Text Summarizer    | HTMX + prompt engineering           | [mb-project-03-summarizer](https://github.com/EOEboh/mb-project-03-summarizer)             |
| 4  | AI Resume Analyzer       | File uploads + PDF extraction       | [mb-project-04-resume-analyzer](https://github.com/EOEboh/mb-project-04-resume-analyzer)   |
| 5  | AI Writing Assistant     | SQLite + contextual AI commands     | [mb-project-05-writing-assistant](https://github.com/EOEboh/mb-project-05-writing-assistant)|
| 6  | Image Caption Generator  | Multimodal (LLaVA vision model)     | [mb-project-06-image-captioner](https://github.com/EOEboh/mb-project-06-image-captioner)   |
| 7  | API Doc Generator        | Multi-pass prompting + export       | [mb-project-07-api-doc-generator](https://github.com/EOEboh/mb-project-07-api-doc-generator)|
| 8  | Meeting Notes Summarizer | Whisper audio pipeline              | [mb-project-08-meeting-notes](https://github.com/EOEboh/mb-project-08-meeting-notes)       |
| 9  | AI Agent with Tool Use   | ReAct pattern + tool dispatch       | [mb-project-09-ai-agent](https://github.com/EOEboh/mb-project-09-ai-agent)                 |
| 10 | Full-Stack AI SaaS       | JWT + multi-tenancy + Docker        | [mb-project-10-ai-saas](https://github.com/EOEboh/mb-project-10-ai-saas)                   |