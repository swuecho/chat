# Running Async Python Code in Pyodide

## Overview

This document outlines the challenges and solutions for executing asynchronous Python code in the Pyodide-based Python runner used in this chat application.

## The Problem

When users write Python code containing `asyncio.run()` calls, they encounter a runtime error:

```
RuntimeError: asyncio.run() cannot be called from a running event loop
```

This occurs because Pyodide already runs within an active asyncio event loop, and `asyncio.run()` attempts to create a new event loop, which is not allowed.

## Root Cause

- **Pyodide Environment**: Pyodide executes Python code within a JavaScript environment that already has an active asyncio event loop
- **asyncio.run() Limitation**: This function is designed to be the main entry point for asyncio programs and cannot be called from within an existing event loop
- **Common Pattern**: Many Python async tutorials and examples use `if __name__ == "__main__": asyncio.run(main())` which doesn't work in Pyodide

## Solution Implementation

### Detection and Transformation

The Python runner now automatically detects and transforms `asyncio.run()` calls:

1. **Pattern Detection**: Code is scanned for `asyncio.run()` calls
2. **Syntax Transformation**: `asyncio.run(func())` is converted to `await func()`
3. **Context Wrapping**: The entire code is wrapped in an async function context
4. **Pyodide Execution**: Uses `pyodide.runPythonAsync()` instead of `pyodide.runPython()`

### Code Transformation Example

**Original Code:**
```python
import asyncio

async def main():
    print("Hello from async!")
    await asyncio.sleep(1)
    print("Done!")

if __name__ == "__main__":
    asyncio.run(main())
```

**Transformed Code:**
```python
import asyncio

async def _execute_main():
    import asyncio

    async def main():
        print("Hello from async!")
        await asyncio.sleep(1)
        print("Done!")

    if __name__ == "__main__":
        await main()

# Execute the main function
await _execute_main()
```

### Key Changes in Implementation

1. **Regex Replacement**: `asyncio.run(([^)]+))` → `await $1`
2. **Async Wrapper**: Entire code wrapped in `async def _execute_main():`
3. **Top-level Await**: Uses `await _execute_main()` for execution
4. **Execution Method**: Uses `pyodide.runPythonAsync()` for async code

## Technical Details

### Why This Works

- **Pyodide Compatibility**: `runPythonAsync()` is designed to handle top-level await statements
- **Event Loop Reuse**: Instead of creating a new event loop, the code runs within Pyodide's existing loop
- **Proper Awaiting**: The async function is properly awaited, preventing "coroutine was never awaited" warnings

### Error Handling

The runner provides informative feedback:
- Detects `asyncio.run()` usage and notifies the user
- Explains the transformation being applied
- Maintains error context for debugging

## Best Practices

### For Users

1. **Avoid `asyncio.run()`**: In Pyodide environments, use async/await directly
2. **Top-level Async**: Write async functions and let the runner handle execution
3. **Error Awareness**: Understand that Pyodide has different async execution patterns

### For Developers

1. **Detection First**: Always scan for problematic patterns before execution
2. **Clear Messaging**: Inform users when code transformations occur
3. **Fallback Strategy**: Use `runPython()` for synchronous code, `runPythonAsync()` for async
4. **Testing**: Test with various async patterns including nested async calls

## Limitations

1. **Complex Patterns**: Very complex async patterns may still require manual adjustment
2. **Performance**: Code transformation adds slight overhead
3. **Debugging**: Transformed code may be harder to debug than original

## Future Improvements

1. **Better Pattern Recognition**: Handle more complex `asyncio.run()` usage patterns
2. **Source Maps**: Maintain mapping between original and transformed code for better debugging
3. **Optimization**: Cache transformation results for repeated code execution
4. **User Education**: Provide more guidance on async Python patterns in Pyodide

## Conclusion

The async Python code execution solution successfully bridges the gap between standard Python async patterns and Pyodide's execution environment. By automatically detecting and transforming `asyncio.run()` calls, users can run async Python code seamlessly without needing to understand the underlying Pyodide constraints.