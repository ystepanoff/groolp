# Groolp (aka Groovy Gulp)

Groolp is a [Gulp](https://gulpjs.com)-like task runner written in Go that streamlines your development 
workflows by automating common tasks—like building, testing, and deploying—with ease. It allows you to 
define tasks using a simple YAML configuration and extend its functionality with dynamic Lua scripts. 
Whether you need a quick one-liner or a complex build pipeline, Groolp offers a powerful and flexible, 
cross-platform solution to manage your project's tasks.

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

### Installation

#### Building from source

Clone the repository and build Groolp:
```bash
git clone https://github.com/ystepanoff/groolp
cd groolp
go build -o groolp cmd/main.go
```

#### Using pre-built binaries

Download the latest release from the [releases page](https://github.com/ystepanoff/groolp/releases) and add it to your PATH.

### Project Setup

1. **Initialize a new project:**
```bash
groolp init
```
This creates a `.groolp` directory with:
- `tasks.yaml` - Your task definitions
- `scripts/` - Directory for Lua scripts
- A sample Lua script to help you get started

2. **Basic task definition in `tasks.yaml`:**
```yaml
tasks:
  build:
    command: go build -o myapp .
    description: Build the application
    watch:
      - "*.go"
      - "go.mod"
      - "go.sum"

  test:
    command: go test ./...
    description: Run tests
    depends:
      - build

  lint:
    command: golangci-lint run
    description: Run linter
    depends:
      - build
```

3. **Advanced tasks with Lua scripts:**
Create a file in `.groolp/scripts/custom_task.lua`:
```lua
function run()
    -- Access persistent storage
    local last_run = get_data("last_run")
    if last_run then
        print("Last run: " .. last_run)
    end
    
    -- Run commands
    local result = run_command("go version")
    print("Go version: " .. result)
    
    -- Store data
    set_data("last_run", os.date())
    
    -- Complex logic
    if some_condition then
        run_command("go build -o myapp .")
    end
end
```

### Common Use Cases

1. **Development Workflow**
```yaml
tasks:
  dev:
    command: go run main.go
    watch:
      - "*.go"
      - "templates/*"
    description: Run development server with hot reload

  format:
    command: go fmt ./...
    description: Format all Go files

  clean:
    command: rm -rf build/
    description: Clean build artifacts
```

2. **Build Pipeline**
```yaml
tasks:
  build-all:
    depends:
      - lint
      - test
      - build
    description: Run complete build pipeline

  build:
    command: |
      go build -o build/app
      cp -r templates build/
    description: Build application with assets

  deploy:
    command: ./scripts/deploy.sh
    depends:
      - build-all
    description: Deploy to production
```

3. **Testing and Quality**
```yaml
tasks:
  test-coverage:
    command: go test -coverprofile=coverage.out ./...
    description: Generate test coverage

  test-html:
    command: go tool cover -html=coverage.out
    depends:
      - test-coverage
    description: View coverage in browser
```

### Configuration Options

#### Task Configuration

Tasks in `tasks.yaml` support the following options:
- `command`: Shell command to execute
- `description`: Human-readable task description
- `depends`: List of task dependencies
- `watch`: List of file patterns to watch for changes
- `script`: Path to Lua script (relative to `.groolp/scripts/`)
- `env`: Environment variables for the task
- `timeout`: Maximum execution time in seconds

Example with all options:
```yaml
tasks:
  complex-task:
    command: ./build.sh
    description: Complex build task
    depends:
      - prepare
      - validate
    watch:
      - "src/**/*"
      - "config/*.yaml"
    script: complex_build.lua
    env:
      GOOS: linux
      GOARCH: amd64
    timeout: 300
```

#### Lua Script API

Available functions in Lua scripts:
- `run_command(cmd)`: Execute shell command and return output
- `get_data(key)`: Retrieve stored data
- `set_data(key, value)`: Store data persistently
- `watch_files(patterns)`: Add file patterns to watch
- `log(message, level)`: Log messages (levels: info, warn, error)

### Best Practices

1. **Task Organization**
   - Group related tasks together
   - Use meaningful task names
   - Document complex tasks with descriptions
   - Keep tasks focused and single-purpose

2. **Performance**
   - Use file watching judiciously
   - Implement proper task dependencies
   - Cache expensive operations using persistent storage
   - Use Lua scripts for complex logic

3. **Maintenance**
   - Keep tasks.yaml clean and well-documented
   - Version control your .groolp directory
   - Use consistent naming conventions
   - Document custom Lua scripts

### Troubleshooting

Common issues and solutions:

1. **Task not running:**
   - Check task dependencies
   - Verify command syntax
   - Ensure required tools are installed
   - Check file permissions

2. **Watch not working:**
   - Verify file patterns in watch section
   - Check file system events are supported
   - Ensure no circular dependencies

3. **Lua script errors:**
   - Check script syntax
   - Verify API function usage
   - Review error messages in logs

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
- Follow Go's [Effective Go](https://go.dev/doc/effective_go) guidelines.
- Ensure your code is formatted (`go fmt`) and linted.
- Write clear and concise commit messages.

### Reporting issues
If you encounter bugs or have feature requests, please open an issue on GitHub with detailed information.

