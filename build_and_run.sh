#!/bin/bash

# --- Configuration ---
GO_FILE="main.go" # Your Go source file
EXEC_NAME="emailserver" # Desired output executable name (without .exe)
CAPABILITY="cap_net_bind_service" # Example capability (e.g., for port 80/443)

# --- Script ---

echo "--- Building Go Executable ---"
# Build the Go program, outputting to current directory
go build -o "$EXEC_NAME" || { echo "Go build failed!"; exit 1; }
echo "Build complete: $EXEC_NAME"

echo "--- Applying Capabilities (Requires sudo) ---"
# Check if we are root, otherwise prompt for sudo
if [ "$EUID" -ne 0 ]; then
  echo "Please run with sudo to apply capabilities."
  sudo setcap "cap_net_bind_service=+eip" "./$EXEC_NAME"
else
  setcap "cap_net_bind_service=+eip" "./$EXEC_NAME"
fi

if [ $? -eq 0 ]; then
  echo "Capabilities set successfully."
else
  echo "Failed to set capabilities. Check sudo or capability name."
  # exit 1 # Uncomment if you want to stop on setcap failure
fi

echo "--- Running Executable ---"
# Run the built executable
./"$EXEC_NAME"

echo "--- Script Finished ---"
