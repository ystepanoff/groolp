# Groolp (aka Groovy Gulp)

Groolp is a [Gulp](https://gulpjs.com)-like task runner written in Go that streamlines your development 
workflows by automating common tasks—like building, testing, and deploying—with ease. It allows you to 
define tasks using a simple YAML configuration and extend its functionality with dynamic Lua scripts. 
Whether you need a quick one-liner or a complex build pipeline, Groolp offers a powerful and flexible, 
cross-platform solution to manage your project’s tasks.

## Features

- **Сross-platform**

Runs seamlessly on Linux, macOS, and Windows.
- **Simple task definition**

Define simple tasks using an easy-to-read `tasks.yaml` file. Just specify the command and any dependencies, 
and Groolp handles the rest. For more advanced tasks, use dynamic Lua scripts; Groolp exposes useful 
bridging functions like `run_command`, `set_data`, and `get_data`.
- **Task dependencies**

Set dependencies in your `tasks.yaml` or Lua scripts to create robust, ordered pipelines. Groolp detects 
circular dependencies and prevents them.
- **File Changes Watcher**

Groolp includes a file watcher that detects changes in your project and can automatically trigger 
tasks. Perfect for continuous integration or live development workflows.
- **Persistent data storage**
Can be accessed from Lua tasks for get/set interactions and is automatically saved in a per-project `.groolp` directory.

## Usage

### Building from source

Clone the repository and build Groolp:
```bash
git clone https://github.com/ystepanoff/groolp
cd groolp
go build -o groolp cmd/main.go
```
This produces a `groolp` binary that you can run from your project root. 

### Bootstrap your project

On first run, Groolp automatically creates a .groolp directory in your current working directory if it doesn’t already exist. This directory includes:
- Sample `tasks.yaml` file for defining simple tasks.
- `scripts` subdirectory for your Lua scripts (containing a sample Lua script to get you started).
- A sample Lua script to get you started.

## Contributions to .groolp

Thank you for considering contributing to Groolp! Contributions of all kinds are welcome.

### How to Contribute

1. **Fork the repository:**
- Click the "Fork" button on GitHub to create your own copy.

2. **Clone your fork:**
```bash
git clone https://github.com/your-username/groolp.git
cd groolp
```

3. **Create a new branch:**
```bash
git checkout -b feature/your-feature-name
```

4. **Make new changes:**
- Implement your feature or fix.

5. **Commit & Push:**
```bash
git commit -m "Add feature: your feature description"
git push origin feature/your-feature-name
```

6. **Submit a pull request**:
- Go to the original Groolp repository and submit a pull request.

### Coding Standards 
- Follow Go’s [Effective Go](https://go.dev/doc/effective_go) guidelines.
- Ensure your code is formatted (`go fmt`) and linted.
- Write clear and concise commit messages.

### Reporting issues
If you encounter bugs or have feature requests, please open an issue on GitHub with detailed information.

