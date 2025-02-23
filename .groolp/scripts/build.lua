-- deloy_pipeline.lua
-- This script registers tasks to clean, build, test, and deploy a project.

-- Task: clean
-- Cleans the build directory.
register_task("clean", "Clean the build directory", function()
	set_data("cleaned", false)
	local code, err = run_command("rm -rf build && mkdir build")
	if err then
		error("Failed to clean build directory: " .. err)
	elseif code ~= 0 then
		error("Clean command returned exit code " .. code)
	else
		print("Build directory cleaned.")
	end
	set_data("cleaned", true)
end)

-- Task: lint
-- Runs golangci-lint
register_task("lint", "Run golangci-lint", function()
	set_data("lint_success", false)
	local code, err = run_command("golangci-lint run --show-stats -v")
	if err then
		error("golangci-lint failed: " .. err)
	elseif code ~= 0 then
		error("golangci-lint exited with code " .. code)
	else
		print("golangci-lint passed successfully!")
	end
	set_data("lint_success", true)
end, { "clean" })

-- Task: build
-- Builds the project; depends on "clean".
register_task("build", "Compile the project source code", function()
	local cleaned = get_data("cleaned")
	if not cleaned then
		error("Build prerequisite not met: build directory not cleaned.")
	end
	local lint_success = get_data("lint_success")
	if not lint_success then
		error("Build preprequisite not met: linting was not successful.")
	end

	local code, err = run_command("go build -o build/groolp ./cmd/main.go")
	if err then
		error("Build failed: " .. err)
	elseif code ~= 0 then
		error("Build command returned exit code " .. code)
	else
		print("Project built successfully.")
	end
	set_data("buildTime", os.date("%Y-%m-%d %H:%M:%S"))
end, { "clean", "lint", "test" })

-- Task: test
-- Runs unit tests; depends on "build".
register_task("test", "Run unit tests for the project", function()
	local code, err = run_command("go test -v -race ./...")
	if err then
		error("Tests failed: " .. err)
	elseif code ~= 0 then
		error("Tests returned exit code " .. code)
	else
		print("All tests passed.")
	end
end, { "clean", "lint" })
