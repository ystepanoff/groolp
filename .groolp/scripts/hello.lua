-- sample.lua
-- Register a sample plugin-based task in Lua

register_task(
  "sample-lua-task",
  "Print a greeting from sample.lua",
  function()
    print("Hello from sample.lua!")
  end
)
