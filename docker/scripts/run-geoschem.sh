#!/bin/bash

# GeosChem Rocky Linux 9 runtime script
set -e

# Source Spack environment
source /opt/spack/share/spack/setup-env.sh
spack env activate geoschem

echo "GeosChem environment activated (Rocky Linux 9)"
echo "Spack environment: $(spack env status)"
echo "Available modules:"
spack find

# If no arguments provided, start interactive shell
if [ $# -eq 0 ]; then
    echo "Starting interactive shell..."
    exec /bin/bash
fi

# Execute the provided command
echo "Executing: $@"
exec "$@"